package api

const (
	tagFields = `
id
name
count
`

	fileFields = `
path
isDir
size
createdAt
updatedAt
children
mediaId
`

	imageFields = `
id
title
path
size
bucketId
takenAt
createdAt
updatedAt
tags {
  ` + tagFields + `
}
`

	videoFields = `
id
title
path
duration
size
bucketId
takenAt
createdAt
updatedAt
tags {
  ` + tagFields + `
}
`

	audioFields = `
id
title
artist
path
duration
size
bucketId
albumFileId
createdAt
updatedAt
tags {
  ` + tagFields + `
}
`

	noteFields = `
id
title
content
deletedAt
createdAt
updatedAt
tags {
  ` + tagFields + `
}
`

	feedFields = `
id
name
url
fetchContent
createdAt
updatedAt
`

	feedEntryFields = `
id
title
url
image
author
description
content
feedId
rawId
publishedAt
createdAt
updatedAt
tags {
  ` + tagFields + `
}
`

	messageFields = `
id
body
address
serviceCenter
date
type
threadId
subscriptionId
isMms
attachments {
  path
  contentType
  name
}
tags {
  ` + tagFields + `
}
`

	messageConversationFields = `
id
address
snippet
date
messageCount
read
`

	contentItemFields = `
label
value
type
`

	contactFields = `
id
prefix
suffix
firstName
middleName
lastName
updatedAt
notes
source
thumbnailId
starred
phoneNumbers {
  label
  value
  type
  normalizedNumber
}
addresses {
  ` + contentItemFields + `
}
emails {
  ` + contentItemFields + `
}
websites {
  ` + contentItemFields + `
}
events {
  ` + contentItemFields + `
}
ims {
  ` + contentItemFields + `
}
tags {
  ` + tagFields + `
}
`

	callFields = `
id
name
number
duration
accountId
startedAt
photoId
type
geo {
  isp
  city
  province
}
tags {
  ` + tagFields + `
}
`

	packageFields = `
id
name
type
version
path
size
certs {
  issuer
  subject
  serialNumber
  validFrom
  validTo
}
installedAt
updatedAt
`

	notificationFields = `
id
onlyOnce
isClearable
appId
appName
time
silent
title
body
actions
replyActions
`

	bookmarkFields = `
id
url
title
faviconPath
groupId
pinned
clickCount
lastClickedAt
sortOrder
createdAt
updatedAt
`

	bookmarkGroupFields = `
id
name
collapsed
sortOrder
createdAt
updatedAt
`

	chatItemFields = `
id
fromId
toId
channelId
createdAt
content
data
`

	chatChannelFields = `
id
name
owner
members {
  id
  status
}
version
status
createdAt
updatedAt
`

	peerFields = `
id
name
ip
status
port
deviceType
createdAt
updatedAt
`

	mountFields = `
id
name
path
mountPoint
fsType
totalBytes
usedBytes
freeBytes
remote
alias
driveType
diskId
`

	playlistAudioFields = `
title
artist
path
duration
`

	appFields = `
usbConnected
urlToken
httpPort
httpsPort
appDir
deviceName
battery
appVersion
osVersion
channel
permissions
audios {
  ` + playlistAudioFields + `
}
audioCurrent
audioMode
sdcardPath
usbDiskPaths
internalStoragePath
downloadsDir
developerMode
favoriteFolders {
  rootPath
  fullPath
  alias
}
`

	deviceInfoFields = `
deviceName
releaseBuildVersion
versionCodeName
manufacturer
securityPatch
bootloader
deviceId
model
product
fingerprint
hardware
radioVersion
device
board
displayVersion
buildBrand
buildHost
buildTime
uptime
buildUser
serial
osVersion
language
sdkVersion
javaVmVersion
kernelVersion
glEsVersion
screenDensity
screenHeight
screenWidth
phoneNumbers {
  id
  name
  number
}
`

	batteryFields = `
level
voltage
health
plugged
temperature
status
technology
capacity
`

	pomodoroTodayFields = `
date
completedCount
currentRound
timeLeft
totalTime
isRunning
isPause
state
`

	pomodoroSettingsFields = `
workDuration
shortBreakDuration
longBreakDuration
pomodorosBeforeLongBreak
showNotification
playSoundOnComplete
`

	imageSearchStatusFields = `
status
downloadProgress
errorMessage
modelSize
modelDir
isIndexing
totalImages
indexedImages
`

	screenMirrorQualityFields = `
mode
resolution
`
)

const (
	queryApp = `query {
  app {
    ` + appFields + `
  }
}`

	queryDeviceInfo = `query {
  deviceInfo {
    ` + deviceInfoFields + `
  }
}`

	queryBattery = `query {
  battery {
    ` + batteryFields + `
  }
}`

	queryPeers = `query {
  peers {
    ` + peerFields + `
  }
}`

	queryMounts = `query {
  mounts {
    ` + mountFields + `
  }
}`

	queryFiles = `query files($root: String!, $offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  files(root: $root, offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    ` + fileFields + `
  }
}`

	queryRecentFiles = `query {
  recentFiles {
    ` + fileFields + `
  }
}`

	queryFileInfo = `query fileInfo($id: ID!, $path: String!, $fileName: String!) {
  fileInfo(id: $id, path: $path, fileName: $fileName) {
    ... on FileInfo {
      path
      updatedAt
      size
      tags {
        id
        name
      }
    }
    data {
      ... on ImageFileInfo {
        width
        height
        location {
          latitude
          longitude
        }
      }
      ... on VideoFileInfo {
        duration
        width
        height
        location {
          latitude
          longitude
        }
      }
      ... on AudioFileInfo {
        duration
        location {
          latitude
          longitude
        }
      }
    }
  }
}`

	queryAppFiles = `query appFiles($offset: Int!, $limit: Int!) {
  appFiles(offset: $offset, limit: $limit) {
    id
    size
    mimeType
    fileName
    createdAt
    updatedAt
  }
  appFileCount
}`

	queryImages = `query images($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  images(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    ` + imageFields + `
  }
  imageCount(query: $query)
}`

	queryVideos = `query videos($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  videos(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    ` + videoFields + `
  }
  videoCount(query: $query)
}`

	queryAudios = `query audios($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  items: audios(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    ` + audioFields + `
  }
  total: audioCount(query: $query)
}`

	queryMediaBuckets = `query mediaBuckets($type: DataType!) {
  mediaBuckets(type: $type) {
    id
    name
    itemCount
    topItems
  }
}`

	queryTags = `query tags($type: DataType!) {
  tags(type: $type) {
    ` + tagFields + `
  }
}`

	querySMS = `query sms($offset: Int!, $limit: Int!, $query: String!) {
  sms(offset: $offset, limit: $limit, query: $query) {
    ` + messageFields + `
  }
  smsCount(query: $query)
}`

	querySMSConversations = `query smsConversations($offset: Int!, $limit: Int!, $query: String!) {
  smsConversations(offset: $offset, limit: $limit, query: $query) {
    ` + messageConversationFields + `
  }
  smsConversationCount(query: $query)
}`

	queryContacts = `query contacts($offset: Int!, $limit: Int!, $query: String!) {
  contacts(offset: $offset, limit: $limit, query: $query) {
    ` + contactFields + `
  }
  contactCount(query: $query)
}`

	queryContactSources = `query {
  contactSources {
    name
    type
  }
}`

	queryCalls = `query calls($offset: Int!, $limit: Int!, $query: String!) {
  calls(offset: $offset, limit: $limit, query: $query) {
    ` + callFields + `
  }
  callCount(query: $query)
}`

	queryNotes = `query notes($offset: Int!, $limit: Int!, $query: String!) {
  notes(offset: $offset, limit: $limit, query: $query) {
    id
    title
    deletedAt
    createdAt
    updatedAt
    tags {
      id
      name
    }
  }
  noteCount(query: $query)
}`

	queryNote = `query note($id: ID!) {
  note(id: $id) {
    ` + noteFields + `
  }
}`

	queryFeeds = `query {
  feeds {
    ` + feedFields + `
  }
}`

	queryFeedEntries = `query feedEntries($offset: Int!, $limit: Int!, $query: String!) {
  items: feedEntries(offset: $offset, limit: $limit, query: $query) {
    id
    title
    url
    image
    author
    feedId
    rawId
    publishedAt
    createdAt
    updatedAt
    tags {
      id
      name
    }
  }
  total: feedEntryCount(query: $query)
}`

	queryFeedEntry = `query feedEntry($id: ID!) {
  feedEntry(id: $id) {
    ` + feedEntryFields + `
    feed {
      ` + feedFields + `
    }
  }
}`

	queryPackages = `query packages($offset: Int!, $limit: Int!, $query: String!, $sortBy: FileSortBy!) {
  packages(offset: $offset, limit: $limit, query: $query, sortBy: $sortBy) {
    ` + packageFields + `
  }
  packageCount(query: $query)
}`

	queryPackageStatuses = `query packageStatuses($ids: [ID!]!) {
  packageStatuses(ids: $ids) {
    id
    exist
    updatedAt
  }
}`

	queryNotifications = `query {
  notifications {
    ` + notificationFields + `
  }
}`

	queryBookmarks = `query {
  bookmarks {
    ` + bookmarkFields + `
  }
}`

	queryBookmarkGroups = `query {
  bookmarkGroups {
    ` + bookmarkGroupFields + `
  }
}`

	queryScreenMirror = `query {
  screenMirrorState
  screenMirrorControlEnabled
  screenMirrorQuality {
    ` + screenMirrorQualityFields + `
  }
}`

	queryPomodoroToday = `query {
  pomodoroToday {
    ` + pomodoroTodayFields + `
  }
}`

	queryPomodoroSettings = `query {
  pomodoroSettings {
    ` + pomodoroSettingsFields + `
  }
}`

	queryChatItems = `query chatItems($id: String!) {
  chatItems(id: $id) {
    ` + chatItemFields + `
  }
}`

	queryChatChannels = `query {
  chatChannels {
    ` + chatChannelFields + `
  }
}`

	queryUploadedChunks = `query uploadedChunks($fileId: String!) {
  uploadedChunks(fileId: $fileId)
}`

	queryImageSearchStatus = `query {
  imageSearchStatus {
    ` + imageSearchStatusFields + `
	}
}`
)

var _ = [...]string{
	tagFields,
	fileFields,
	imageFields,
	videoFields,
	audioFields,
	noteFields,
	feedFields,
	feedEntryFields,
	messageFields,
	messageConversationFields,
	contentItemFields,
	contactFields,
	callFields,
	packageFields,
	notificationFields,
	bookmarkFields,
	bookmarkGroupFields,
	chatItemFields,
	chatChannelFields,
	peerFields,
	mountFields,
	playlistAudioFields,
	appFields,
	deviceInfoFields,
	batteryFields,
	pomodoroTodayFields,
	pomodoroSettingsFields,
	imageSearchStatusFields,
	screenMirrorQualityFields,
	queryApp,
	queryDeviceInfo,
	queryBattery,
	queryPeers,
	queryMounts,
	queryFiles,
	queryRecentFiles,
	queryFileInfo,
	queryAppFiles,
	queryImages,
	queryVideos,
	queryAudios,
	queryMediaBuckets,
	queryTags,
	querySMS,
	querySMSConversations,
	queryContacts,
	queryContactSources,
	queryCalls,
	queryNotes,
	queryNote,
	queryFeeds,
	queryFeedEntries,
	queryFeedEntry,
	queryPackages,
	queryPackageStatuses,
	queryNotifications,
	queryBookmarks,
	queryBookmarkGroups,
	queryScreenMirror,
	queryPomodoroToday,
	queryPomodoroSettings,
	queryChatItems,
	queryChatChannels,
	queryUploadedChunks,
	queryImageSearchStatus,
}
