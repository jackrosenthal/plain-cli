# Plain App API Reference

Self-contained reference for communicating with the Plain backend. No source files needed.

---

## Transports

| Transport | Endpoint | Purpose |
|-----------|----------|---------|
| Encrypted GraphQL / HTTP | `POST /graphql` | All data queries & mutations |
| Encrypted WebSocket | `GET /` (upgrade) | Real-time push events |
| REST | `POST /init`, `GET /health_check` | Auth & diagnostics |
| REST | `GET /fs`, `GET /proxyfs` | File access |
| REST | `POST /upload`, `POST /upload_chunk` | File upload |

Base URL is determined by the host the app is served from. WebSocket uses the same host with `ws://` or `wss://` depending on whether the HTTP connection is plain or TLS.

---

## Encryption

**Every** request and response body is end-to-end encrypted. The server never sees plaintext.

### Algorithm

**XChaCha20-Poly1305**

Wire format for any encrypted blob:
```
[24 bytes: random nonce][ciphertext + 16-byte Poly1305 auth tag]
```

### Key derivation

There are two key sources depending on context:

**Session key** (used for all GraphQL, WebSocket app traffic, and file ID encryption):
```
key = base64_decode(auth_token)   // first 32 bytes used
```

**Login key** (used only during the authentication WebSocket handshake):
```
hash = sha512_hex(password)
key  = ascii_bytes(hash[0:32])    // first 32 ASCII chars of hex string → 32-byte key
```

### Replay protection (GraphQL only)

Before encrypting a GraphQL payload, it is wrapped:
```
"{server_timestamp}|{16_char_random_hex_nonce}|{graphql_json}"
```

`server_timestamp` is a Unix millisecond timestamp synchronized to the server clock via `window.__SERVER_TIME__` (injected by the server at page load).

---

## Required Header

All requests must include:

```
c-id: <client_id>
```

`client_id` is a stable UUID generated once per browser and persisted in `localStorage`. It identifies the client session.

---

## Authentication Flow

### Step 1 — Init check

```
POST /init
Content-Type: multipart/form-data
c-id: <client_id>

Body (if auth_token exists):
  encrypt(session_key, random_uuid_string)   // binary ciphertext
Body (if no auth_token):
  (empty)
```

**Responses:**

| Status | Body | Meaning |
|--------|------|---------|
| 403 | any | Web access disabled on this device |
| 200 | empty | Token is valid → client should redirect to app |
| 200 | string | Pre-filled password (automated login) |

### Step 2 — Password authentication (WebSocket)

Open a WebSocket to:
```
ws[s]://<host>?cid=<client_id>&auth=1
```

On open, send one binary frame:
```
encrypt(login_key, json_string)
```

Where `json_string` is:
```json
{
  "password": "<sha512_hex_of_password>",
  "browserName": "Chrome",
  "browserVersion": "120",
  "osName": "Windows",
  "osVersion": "10",
  "isMobile": false
}
```

**Server responses** (binary frames, decrypt with `login_key`):

```json
{ "status": "PENDING" }
```
→ User must confirm on the device.

```json
{ "token": "<base64_auth_token>" }
```
→ Authentication succeeded. Store `token` as `auth_token` and use as session key.

If the WebSocket closes with a non-empty `reason` string, that reason is the error code (e.g. `"wrong_password"`, `"expired"`).

### Step 3 — Health check (diagnostic only)

```
GET /health_check
```

Returns 200 if the HTTP server is reachable. Used to distinguish "server down" from "WebSocket failed."

---

## GraphQL Endpoint

```
POST /graphql
Content-Type: multipart/form-data
c-id: <client_id>

Body: binary ciphertext
  = encrypt(session_key, replay_wrapped_graphql_json)
```

Where `replay_wrapped_graphql_json` is:
```
"<unix_ms>|<16_char_hex_nonce>|<standard_graphql_json>"
```

And `standard_graphql_json` is a normal GraphQL request body:
```json
{
  "query": "query { ... }",
  "variables": { ... }
}
```

**Response:** binary ciphertext. Decrypt with `session_key` to get standard GraphQL JSON response.

**HTTP error handling:**

| Status | Action |
|--------|--------|
| 401 | Session expired — discard `auth_token`, re-authenticate |
| 403 | Web access disabled — surface error to user |
| (timeout after 30 s) | Treat as `connection_timeout` error |

---

## GraphQL Schema

### Common types

```graphql
type Tag { id: ID!, name: String!, count: Int }

type File {
  path: String!, isDir: Boolean!, size: Int!
  createdAt: String!, updatedAt: String!
  children: Int, mediaId: String
}

type Image {
  id: ID!, title: String, path: String!, size: Int!
  bucketId: String, takenAt: String
  createdAt: String!, updatedAt: String!
  tags: [Tag!]!
}

type Video {
  id: ID!, title: String, path: String!, duration: Int
  size: Int!, bucketId: String, takenAt: String
  createdAt: String!, updatedAt: String!
  tags: [Tag!]!
}

type Audio {
  id: ID!, title: String, artist: String, path: String!
  duration: Int, size: Int!, bucketId: String, albumFileId: String
  createdAt: String!, updatedAt: String!
  tags: [Tag!]!
}

type Note {
  id: ID!, title: String!, content: String!
  deletedAt: String, createdAt: String!, updatedAt: String!
  tags: [Tag!]!
}

type Feed { id: ID!, name: String!, url: String!, fetchContent: Boolean!, createdAt: String!, updatedAt: String! }

type FeedEntry {
  id: ID!, title: String!, url: String!, image: String, author: String
  description: String, content: String, feedId: ID!, rawId: String
  publishedAt: String, createdAt: String!, updatedAt: String!
  tags: [Tag!]!
}

type Message {
  id: ID!, body: String!, address: String!, serviceCenter: String
  date: String!, type: Int!, threadId: String!, subscriptionId: String
  isMms: Boolean!
  attachments: [{ path: String!, contentType: String!, name: String }]
  tags: [Tag!]!
}

type MessageConversation {
  id: ID!, address: String!, snippet: String, date: String!
  messageCount: Int!, read: Boolean!
}

type Contact {
  id: ID!, prefix: String, suffix: String
  firstName: String, middleName: String, lastName: String
  updatedAt: String!, notes: String, source: String
  thumbnailId: String, starred: Boolean!
  phoneNumbers: [{ label: String, value: String!, type: String, normalizedNumber: String }]
  addresses: [ContentItem], emails: [ContentItem]
  websites: [ContentItem], events: [ContentItem], ims: [ContentItem]
  tags: [Tag!]!
}

type ContentItem { label: String, value: String!, type: String }

type Call {
  id: ID!, name: String, number: String!, duration: Int
  accountId: String, startedAt: String!, photoId: String
  type: Int!  # 1=incoming 2=outgoing 3=missed
  geo: { isp: String, city: String, province: String }
  tags: [Tag!]!
}

type Package {
  id: ID!, name: String!, type: String!, version: String!
  path: String!, size: Int!
  certs: [{ issuer: String, subject: String, serialNumber: String, validFrom: String, validTo: String }]
  installedAt: String!, updatedAt: String!
}

type Notification {
  id: ID!, onlyOnce: Boolean, isClearable: Boolean
  appId: String!, appName: String!, time: String!
  silent: Boolean, title: String, body: String
  actions: [String], replyActions: [String]
}

type Bookmark {
  id: ID!, url: String!, title: String, faviconPath: String
  groupId: String!, pinned: Boolean!, clickCount: Int!
  lastClickedAt: String, sortOrder: Int!
  createdAt: String!, updatedAt: String!
}

type BookmarkGroup {
  id: ID!, name: String!, collapsed: Boolean!, sortOrder: Int!
  createdAt: String!, updatedAt: String!
}

type ChatItem {
  id: ID!, fromId: String!, toId: String!, channelId: String
  createdAt: String!, content: String!
  # content is a JSON string; parse it client-side
  data: MessageImages | MessageFiles | MessageText
}
type MessageImages { ids: [String!]! }
type MessageFiles  { ids: [String!]! }
type MessageText   { ids: [String!]! }

type ChatChannel {
  id: ID!, name: String!, owner: String!
  members: [{ id: String!, status: String! }]
  version: Int!, status: String!
  createdAt: String!, updatedAt: String!
}

type Peer { id: ID!, name: String!, ip: String!, status: String!, port: Int!, deviceType: String!, createdAt: String!, updatedAt: String! }

type Mount {
  id: ID!, name: String, path: String!, mountPoint: String!
  fsType: String, totalBytes: Int!, usedBytes: Int, freeBytes: Int!
  remote: Boolean, alias: String, driveType: String, diskID: String
}

type App {
  usbConnected: Boolean!, urlToken: String, httpPort: Int!, httpsPort: Int
  appDir: String!, deviceName: String!, battery: Int!, appVersion: String!
  osVersion: String!, channel: String!, permissions: [String!]!
  audios: [PlaylistAudio!]!, audioCurrent: String, audioMode: String
  sdcardPath: String, usbDiskPaths: [String!]!, internalStoragePath: String!
  downloadsDir: String!, developerMode: Boolean!
  favoriteFolders: [{ rootPath: String!, fullPath: String!, alias: String }]
}

type PlaylistAudio { title: String, artist: String, path: String!, duration: Int }

type DeviceInfo {
  deviceName: String!, releaseBuildVersion: String, versionCodeName: String
  manufacturer: String!, securityPatch: String, bootloader: String
  deviceId: String!, model: String!, product: String, fingerprint: String
  hardware: String, radioVersion: String, device: String, board: String
  displayVersion: String, buildBrand: String, buildHost: String
  buildTime: String, uptime: String, buildUser: String, serial: String
  osVersion: String!, language: String, sdkVersion: String
  javaVmVersion: String, kernelVersion: String, glEsVersion: String
  screenDensity: Int, screenHeight: Int, screenWidth: Int
  phoneNumbers: [{ id: String, name: String, number: String }]
}

type Battery {
  level: Int!, voltage: Int, health: String, plugged: Boolean!
  temperature: Float, status: String, technology: String, capacity: Int
}

type ScreenMirrorQuality { mode: ScreenMirrorMode!, resolution: String }

type PomodoroToday {
  date: String!, completedCount: Int!, currentRound: Int!
  timeLeft: Int!, totalTime: Int!, isRunning: Boolean!
  isPause: Boolean!, state: String!
}

type PomodoroSettings {
  workDuration: Int!, shortBreakDuration: Int!, longBreakDuration: Int!
  pomodorosBeforeLongBreak: Int!, showNotification: Boolean!, playSoundOnComplete: Boolean!
}

type ImageSearchStatus {
  status: String!, downloadProgress: Float, errorMessage: String
  modelSize: Int, modelDir: String, isIndexing: Boolean!
  totalImages: Int!, indexedImages: Int!
}

enum DataType { IMAGE VIDEO AUDIO NOTE FEED_ENTRY CALL CONTACT MESSAGE BOOKMARK }
enum FileSortBy { NAME NAME_DESC SIZE SIZE_DESC DATE DATE_DESC }
enum MediaPlayMode { ORDER SHUFFLE REPEAT REPEAT_ONE }
enum ScreenMirrorMode { AUTO HD SMOOTH }
```

### Queries

#### Device & app state

```graphql
query { app { ...App } }
query { deviceInfo { ...DeviceInfo }  battery { ...Battery } }
query { peers { id name ip status port deviceType createdAt updatedAt } }
query { mounts { ...Mount } }
query {
  # Home dashboard counts
  smsCount(query: "") contactCount(query: "") callCount(query: "")
  imageCount(query: "") audioCount(query: "") videoCount(query: "")
  packageCount(query: "") noteCount(query: "") feedEntryCount(query: "")
  mounts { id path mountPoint totalBytes freeBytes driveType }
}
```

#### Files

```graphql
query files($root: String!, $offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  files(root: $root, offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) { ...File }
}
query { recentFiles { ...File } }
query fileInfo($id: ID!, $path: String!, $fileName: String!) {
  fileInfo(id: $id, path: $path, fileName: $fileName) {
    ... on FileInfo { path updatedAt size tags { id name } }
    data {
      ... on ImageFileInfo { width height location { latitude longitude } }
      ... on VideoFileInfo { duration width height location { latitude longitude } }
      ... on AudioFileInfo { duration location { latitude longitude } }
    }
  }
}
query appFiles($offset: Int!, $limit: Int!) {
  appFiles(offset: $offset, limit: $limit) { id size mimeType fileName createdAt updatedAt }
  appFileCount
}
```

#### Media

```graphql
query images($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  images(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) { ...Image }
  imageCount(query: $query)
}
query videos($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  videos(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) { ...Video }
  videoCount(query: $query)
}
query audios($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  items: audios(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) { ...Audio }
  total: audioCount(query: $query)
}
query { imageCount(query: "")  imageCount(query: "trash:true") }
query mediaBuckets($type: DataType!) {
  mediaBuckets(type: $type) { id name itemCount topItems }
}
```

#### Tags

```graphql
query tags($type: DataType!) { tags(type: $type) { id name count } }
```

#### Messaging

```graphql
query sms($offset: Int!, $limit: Int!, $query: String!) {
  sms(offset: $offset, limit: $limit, query: $query) { ...Message }
  smsCount(query: $query)
}
query smsConversations($offset: Int!, $limit: Int!, $query: String!) {
  smsConversations(offset: $offset, limit: $limit, query: $query) { ...MessageConversation }
  smsConversationCount(query: $query)
}
query contacts($offset: Int!, $limit: Int!, $query: String!) {
  contacts(offset: $offset, limit: $limit, query: $query) { ...Contact }
  contactCount(query: $query)
}
query { contactSources { name type } }
query calls($offset: Int!, $limit: Int!, $query: String!) {
  calls(offset: $offset, limit: $limit, query: $query) { ...Call }
  callCount(query: $query)
}
```

#### Notes

```graphql
query notes($offset: Int!, $limit: Int!, $query: String!) {
  notes(offset: $offset, limit: $limit, query: $query) { id title deletedAt createdAt updatedAt tags { id name } }
  noteCount(query: $query)
}
query note($id: ID!) { note(id: $id) { ...Note } }
```

#### Feeds

```graphql
query { feeds { ...Feed } }
query feedEntries($offset: Int!, $limit: Int!, $query: String!) {
  items: feedEntries(offset: $offset, limit: $limit, query: $query) {
    id title url image author feedId rawId publishedAt createdAt updatedAt tags { id name }
  }
  total: feedEntryCount(query: $query)
}
query feedEntry($id: ID!) {
  feedEntry(id: $id) { ...FeedEntry  feed { ...Feed } }
}
query {
  total: feedEntryCount(query: "")
  today: feedEntryCount(query: "today:true")
  feedsCount { id count }
}
```

#### Packages

```graphql
query packages($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  packages(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) { ...Package }
  packageCount(query: $query)
}
query packageStatuses($ids: [ID!]!) {
  packageStatuses(ids: $ids) { id exist updatedAt }
}
```

#### Notifications

```graphql
query { notifications { ...Notification } }
```

#### Bookmarks

```graphql
query {
  bookmarks { ...Bookmark }
  bookmarkGroups { ...BookmarkGroup }
}
```

#### Screen mirror

```graphql
query {
  screenMirrorState         # String enum: IDLE | STARTED | ...
  screenMirrorControlEnabled
  screenMirrorQuality { mode resolution }
}
```

#### Pomodoro

```graphql
query { pomodoroSettings { ...PomodoroSettings } }
query {
  pomodoroToday { ...PomodoroToday }
  pomodoroSettings { ...PomodoroSettings }
}
```

#### Chat

```graphql
query chatItems($id: String!) { chatItems(id: $id) { ...ChatItem } }
query { chatChannels { ...ChatChannel } }
```

#### Upload resume

```graphql
query uploadedChunks($fileId: String!) { uploadedChunks(fileId: $fileId) }
# Returns: [Int!] — indices of chunks already received by the server
```

#### Image search

```graphql
query { imageSearchStatus { ...ImageSearchStatus } }
```

---

### Mutations

#### Chat

```graphql
mutation sendChatItem($toId: String!, $content: String!) { sendChatItem(toId: $toId, content: $content) { ...ChatItem } }
mutation deleteChatItem($id: ID!) { deleteChatItem(id: $id) }
mutation createChatChannel($name: String!) { createChatChannel(name: $name) { ...ChatChannel } }
mutation updateChatChannel($id: ID!, $name: String!) { updateChatChannel(id: $id, name: $name) { ...ChatChannel } }
mutation deleteChatChannel($id: ID!) { deleteChatChannel(id: $id) }
mutation leaveChatChannel($id: ID!) { leaveChatChannel(id: $id) }
mutation addChatChannelMember($id: ID!, $peerId: String!) { addChatChannelMember(id: $id, peerId: $peerId) { ...ChatChannel } }
mutation removeChatChannelMember($id: ID!, $peerId: String!) { removeChatChannelMember(id: $id, peerId: $peerId) { ...ChatChannel } }
mutation acceptChatChannelInvite($id: ID!) { acceptChatChannelInvite(id: $id) }
mutation declineChatChannelInvite($id: ID!) { declineChatChannelInvite(id: $id) }
```

#### Files & storage

```graphql
mutation createDir($path: String!) { createDir(path: $path) { ...File } }
mutation writeTextFile($path: String!, $content: String!, $overwrite: Boolean!) { writeTextFile(path: $path, content: $content, overwrite: $overwrite) { ...File } }
mutation renameFile($path: String!, $name: String!) { renameFile(path: $path, name: $name) }
mutation copyFile($src: String!, $dst: String!, $overwrite: Boolean!) { copyFile(src: $src, dst: $dst, overwrite: $overwrite) }
mutation moveFile($src: String!, $dst: String!, $overwrite: Boolean!) { moveFile(src: $src, dst: $dst, overwrite: $overwrite) }
mutation deleteFiles($paths: [String!]!) { deleteFiles(paths: $paths) }
# After uploading all chunks via /upload_chunk:
mutation mergeChunks($fileId: String!, $totalChunks: Int!, $path: String!, $replace: Boolean!, $isAppFile: Boolean!) {
  mergeChunks(fileId: $fileId, totalChunks: $totalChunks, path: $path, replace: $replace, isAppFile: $isAppFile)
}
mutation addFavoriteFolder($rootPath: String!, $fullPath: String!) { addFavoriteFolder(rootPath: $rootPath, fullPath: $fullPath) { rootPath fullPath } }
mutation removeFavoriteFolder($fullPath: String!) { removeFavoriteFolder(fullPath: $fullPath) { rootPath fullPath alias } }
mutation setFavoriteFolderAlias($fullPath: String!, $alias: String!) { setFavoriteFolderAlias(fullPath: $fullPath, alias: $alias) { rootPath fullPath alias } }
```

#### Audio playback

```graphql
mutation playAudio($path: String!) { playAudio(path: $path) { title artist path duration } }
mutation updateAudioPlayMode($mode: MediaPlayMode!) { updateAudioPlayMode(mode: $mode) }
mutation deletePlaylistAudio($path: String!) { deletePlaylistAudio(path: $path) }
mutation addPlaylistAudios($query: String!) { addPlaylistAudios(query: $query) }
mutation clearAudioPlaylist { clearAudioPlaylist }
mutation reorderPlaylistAudios($paths: [String!]!) { reorderPlaylistAudios(paths: $paths) }
```

#### Media management

```graphql
mutation deleteMediaItems($type: DataType!, $query: String!) { deleteMediaItems(type: $type, query: $query) { type query } }
mutation trashMediaItems($type: DataType!, $query: String!)  { trashMediaItems(type: $type, query: $query) { type query } }
mutation restoreMediaItems($type: DataType!, $query: String!) { restoreMediaItems(type: $type, query: $query) { type query } }
```

#### Tags

```graphql
mutation createTag($type: DataType!, $name: String!) { createTag(type: $type, name: $name) { id name count } }
mutation updateTag($id: ID!, $name: String!) { updateTag(id: $id, name: $name) { id name count } }
mutation deleteTag($id: ID!) { deleteTag(id: $id) }
mutation addToTags($type: DataType!, $tagIds: [ID!]!, $query: String!) { addToTags(type: $type, tagIds: $tagIds, query: $query) }
mutation removeFromTags($type: DataType!, $tagIds: [ID!]!, $query: String!) { removeFromTags(type: $type, tagIds: $tagIds, query: $query) }
mutation updateTagRelations($type: DataType!, $item: TagRelationStub!, $addTagIds: [ID!]!, $removeTagIds: [ID!]!) {
  updateTagRelations(type: $type, item: $item, addTagIds: $addTagIds, removeTagIds: $removeTagIds)
}
```

#### Notes

```graphql
input NoteInput { title: String!, content: String! }
mutation saveNote($id: ID!, $input: NoteInput!) { saveNote(id: $id, input: $input) { ...Note } }
mutation deleteNotes($query: String!) { deleteNotes(query: $query) }
mutation trashNotes($query: String!) { trashNotes(query: $query) }
mutation restoreNotes($query: String!) { restoreNotes(query: $query) }
mutation exportNotes($query: String!) { exportNotes(query: $query) }
```

#### Feeds

```graphql
mutation createFeed($url: String!, $fetchContent: Boolean!) { createFeed(url: $url, fetchContent: $fetchContent) { ...Feed } }
mutation updateFeed($id: ID!, $name: String!, $fetchContent: Boolean!) { updateFeed(id: $id, name: $name, fetchContent: $fetchContent) { ...Feed } }
mutation deleteFeed($id: ID!) { deleteFeed(id: $id) }
mutation syncFeeds($id: ID) { syncFeeds(id: $id) }        # id=null syncs all feeds
mutation syncFeedContent($id: ID!) { syncFeedContent(id: $id) { ...FeedEntry  feed { ...Feed } } }
mutation importFeeds($content: String!) { importFeeds(content: $content) }   # OPML XML string
mutation exportFeeds { exportFeeds }                                          # Returns OPML string
mutation deleteFeedEntries($query: String!) { deleteFeedEntries(query: $query) }
mutation saveFeedEntriesToNotes($query: String!) { saveFeedEntriesToNotes(query: $query) }
```

#### Messaging & calls

```graphql
mutation sendSms($number: String!, $body: String!) { sendSms(number: $number, body: $body) }
mutation sendMms($number: String!, $body: String!, $attachmentPaths: [String!]!, $threadId: String!) {
  sendMms(number: $number, body: $body, attachmentPaths: $attachmentPaths, threadId: $threadId)
}
mutation call($number: String!) { call(number: $number) }
mutation deleteCalls($query: String!) { deleteCalls(query: $query) }
mutation deleteContacts($query: String!) { deleteContacts(query: $query) }
mutation setClip($text: String!) { setClip(text: $text) }   # Set device clipboard
```

#### Packages

```graphql
mutation installPackage($path: String!) { installPackage(path: $path) { packageName updatedAt isNew } }
mutation uninstallPackages($ids: [ID!]!) { uninstallPackages(ids: $ids) }
```

#### Notifications

```graphql
mutation cancelNotifications($ids: [ID!]!) { cancelNotifications(ids: $ids) }
mutation replyNotification($id: ID!, $actionIndex: Int!, $text: String!) { replyNotification(id: $id, actionIndex: $actionIndex, text: $text) }
```

#### Bookmarks

```graphql
input BookmarkInput { url: String!, title: String, groupId: String!, pinned: Boolean!, sortOrder: Int! }
mutation addBookmarks($urls: [String!]!, $groupId: String!) { addBookmarks(urls: $urls, groupId: $groupId) { ...Bookmark } }
mutation updateBookmark($id: ID!, $input: BookmarkInput!) { updateBookmark(id: $id, input: $input) { ...Bookmark } }
mutation deleteBookmarks($ids: [ID!]!) { deleteBookmarks(ids: $ids) }
mutation recordBookmarkClick($id: ID!) { recordBookmarkClick(id: $id) }
mutation createBookmarkGroup($name: String!) { createBookmarkGroup(name: $name) { ...BookmarkGroup } }
mutation updateBookmarkGroup($id: ID!, $name: String!, $collapsed: Boolean!, $sortOrder: Int!) {
  updateBookmarkGroup(id: $id, name: $name, collapsed: $collapsed, sortOrder: $sortOrder) { ...BookmarkGroup }
}
mutation deleteBookmarkGroup($id: ID!) { deleteBookmarkGroup(id: $id) }
```

#### Screen mirror

```graphql
mutation startScreenMirror($audio: Boolean!) { startScreenMirror(audio: $audio) }
mutation stopScreenMirror { stopScreenMirror }
mutation requestScreenMirrorAudio { requestScreenMirrorAudio }
mutation updateScreenMirrorQuality($mode: ScreenMirrorMode!) { updateScreenMirrorQuality(mode: $mode) }
mutation sendScreenMirrorControl($input: ScreenMirrorControlInput!) { sendScreenMirrorControl(input: $input) }
```

#### WebRTC signaling

```graphql
input WebRtcSignalingMessage {
  type: String!   # "offer" | "answer" | "candidate"
  sdp: String
  candidate: String
  sdpMid: String
  sdpMLineIndex: Int
}
mutation sendWebRtcSignaling($payload: WebRtcSignalingMessage!) { sendWebRtcSignaling(payload: $payload) }
```

#### Pomodoro

```graphql
mutation startPomodoro($timeLeft: Int!) { startPomodoro(timeLeft: $timeLeft) }
mutation stopPomodoro { stopPomodoro }
mutation pausePomodoro { pausePomodoro }
```

#### Image search (AI)

```graphql
mutation { enableImageSearch }
mutation { disableImageSearch }
mutation { cancelImageDownload }
mutation startImageIndex($force: Boolean) { startImageIndex(force: $force) }
mutation { cancelImageIndex }
```

#### Misc

```graphql
mutation setTempValue($key: String!, $value: String!) { setTempValue(key: $key, value: $value) { key value } }
mutation relaunchApp { relaunchApp }
```

---

## File Access REST Endpoints

### `GET /fs?id={fileId}`

Stream or serve a file. `fileId` is a base64-encoded ciphertext:

```
fileId = base64( encrypt(session_key, path_string) )
      or base64( encrypt(session_key, json_string) )  # when mediaId is needed
```

JSON form (when `mediaId` is present):
```json
{ "path": "/sdcard/DCIM/photo.jpg", "mediaId": "42" }
```

Append `&dl=1` to force a `Content-Disposition: attachment` download instead of inline streaming.

### `GET /proxyfs?id={encryptedPeerUrl}`

Proxies a file from a peer device through the local server. The `id` parameter is:

```
id = base64( encrypt(session_key, peer_https_url) )
```

Where `peer_https_url` is the full peer `/fs` URL, e.g.:
```
https://192.168.1.5:8080/fs?id=<peer_file_id>
```

This avoids CORS and TLS certificate issues when accessing peer devices.

---

## File Upload

### Single chunk (`POST /upload`)

For small files. Body is `multipart/form-data` with a single `file` field.

### Chunked upload (`POST /upload_chunk`)

```
POST /upload_chunk
Content-Type: multipart/form-data
c-id: <client_id>

FormData fields:
  info  → binary ciphertext: encrypt(session_key, json_string)
  file  → Blob (raw bytes for this chunk)
```

Where `json_string` is:
```json
{ "fileId": "<file_id>", "index": 0, "size": 1048576 }
```

- `fileId` — stable identifier for the upload session, derived from file metadata + SHA-256 hash.
- `index` — zero-based chunk index.
- `size` — byte count of this chunk.

Typical chunk size: ~1 MB. Up to 3 chunks upload in parallel.

After all chunks are confirmed (via `uploadedChunks` query), call the `mergeChunks` GraphQL mutation to assemble the file on the device.

---

## WebSocket — Real-time Events

### Connection

```
GET ws[s]://<host>?cid=<client_id>
Upgrade: websocket
```

Immediately after the socket opens, the client sends one binary frame:
```
encrypt(session_key, unix_ms_timestamp_string)
```

This synchronizes clocks for replay protection.

### Frame format (server → client)

Every message is a binary frame:
```
[1 byte: event type][remaining bytes: ciphertext]
```

Decrypt the ciphertext with the session key to get a UTF-8 JSON string. Parse it for the event payload.

### Event types

| Byte | Event | Payload shape |
|------|-------|---------------|
| 1 | `message_created` | `Message` object |
| 2 | `message_deleted` | `{ id: String }` |
| 3 | `message_updated` | `Message` object |
| 4 | `feeds_fetched` | `null` |
| 5 | `screen_mirroring` | `{ state: String }` |
| 6 | `webrtc_signaling` | `WebRtcSignalingMessage` |
| 7 | `notification_created` | `Notification` object |
| 8 | `notification_updated` | `Notification` object |
| 9 | `notification_deleted` | `{ id: String }` |
| 10 | `notification_refreshed` | `null` |
| 11 | `pomodoro_action` | `PomodoroToday` object |
| 12 | `pomodoro_settings_update` | `PomodoroSettings` object |
| 13 | `message_cleared` | `{ threadId: String }` |
| 14 | `screen_mirror_audio_granted` | `null` |
| 15 | `bookmark_updated` | `Bookmark` object |
| 16 | `download_progress` | `{ id: String, progress: Float, done: Boolean }` |
| 18 | `channels_updated` | `ChatChannel[]` |
| 19 | `image_search_updated` | `ImageSearchStatus` object |

### Reconnection

If the socket closes or errors, reconnect with exponential backoff starting at 1 s, increasing by 1 s per attempt, capped at 5 s.

---

## WebRTC — Screen Mirroring

WebRTC signaling rides on the GraphQL + WebSocket transports already described.

### Flow

1. Call `startScreenMirror(audio: true|false)` mutation.
2. Device sends an SDP offer via WebSocket event type 6 (`webrtc_signaling`), payload:
   ```json
   { "type": "offer", "sdp": "v=0\r\n..." }
   ```
3. Client creates a peer connection (receive-only), sets the remote description, generates an SDP answer, and sends it back:
   ```graphql
   mutation { sendWebRtcSignaling(payload: { type: "answer", sdp: "..." }) }
   ```
4. Both sides exchange ICE candidates the same way (type `"candidate"` with `candidate`, `sdpMid`, `sdpMLineIndex` fields).
5. Once the ICE connection is established, the device streams audio/video to the browser.
6. Call `stopScreenMirror` mutation to end the session.

Quality can be changed mid-session with `updateScreenMirrorQuality(mode: AUTO|HD|SMOOTH)`.
