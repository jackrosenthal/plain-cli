package api

const (
	mutationSendChatItem = `mutation sendChatItem($toId: String!, $content: String!) {
  sendChatItem(toId: $toId, content: $content) {
    ` + chatItemFields + `
  }
}`

	mutationDeleteChatItem = `mutation deleteChatItem($id: ID!) {
  deleteChatItem(id: $id)
}`

	mutationCreateChatChannel = `mutation createChatChannel($name: String!) {
  createChatChannel(name: $name) {
    ` + chatChannelFields + `
  }
}`

	mutationUpdateChatChannel = `mutation updateChatChannel($id: ID!, $name: String!) {
  updateChatChannel(id: $id, name: $name) {
    ` + chatChannelFields + `
  }
}`

	mutationDeleteChatChannel = `mutation deleteChatChannel($id: ID!) {
  deleteChatChannel(id: $id)
}`

	mutationLeaveChatChannel = `mutation leaveChatChannel($id: ID!) {
  leaveChatChannel(id: $id)
}`

	mutationAddChatChannelMember = `mutation addChatChannelMember($id: ID!, $peerId: String!) {
  addChatChannelMember(id: $id, peerId: $peerId) {
    ` + chatChannelFields + `
  }
}`

	mutationRemoveChatChannelMember = `mutation removeChatChannelMember($id: ID!, $peerId: String!) {
  removeChatChannelMember(id: $id, peerId: $peerId) {
    ` + chatChannelFields + `
  }
}`

	mutationAcceptChatChannelInvite = `mutation acceptChatChannelInvite($id: ID!) {
  acceptChatChannelInvite(id: $id)
}`

	mutationDeclineChatChannelInvite = `mutation declineChatChannelInvite($id: ID!) {
  declineChatChannelInvite(id: $id)
}`

	mutationCreateDir = `mutation createDir($path: String!) {
  createDir(path: $path) {
    ` + fileFields + `
  }
}`

	mutationWriteTextFile = `mutation writeTextFile($path: String!, $content: String!, $overwrite: Boolean!) {
  writeTextFile(path: $path, content: $content, overwrite: $overwrite) {
    ` + fileFields + `
  }
}`

	mutationRenameFile = `mutation renameFile($path: String!, $name: String!) {
  renameFile(path: $path, name: $name)
}`

	mutationCopyFile = `mutation copyFile($src: String!, $dst: String!, $overwrite: Boolean!) {
  copyFile(src: $src, dst: $dst, overwrite: $overwrite)
}`

	mutationMoveFile = `mutation moveFile($src: String!, $dst: String!, $overwrite: Boolean!) {
  moveFile(src: $src, dst: $dst, overwrite: $overwrite)
}`

	mutationDeleteFiles = `mutation deleteFiles($paths: [String!]!) {
  deleteFiles(paths: $paths)
}`

	mutationMergeChunks = `mutation mergeChunks($fileId: String!, $totalChunks: Int!, $path: String!, $replace: Boolean!, $isAppFile: Boolean!) {
  mergeChunks(fileId: $fileId, totalChunks: $totalChunks, path: $path, replace: $replace, isAppFile: $isAppFile)
}`

	mutationAddFavoriteFolder = `mutation addFavoriteFolder($rootPath: String!, $fullPath: String!) {
  addFavoriteFolder(rootPath: $rootPath, fullPath: $fullPath) {
    rootPath
    fullPath
  }
}`

	mutationRemoveFavoriteFolder = `mutation removeFavoriteFolder($fullPath: String!) {
  removeFavoriteFolder(fullPath: $fullPath) {
    rootPath
    fullPath
    alias
  }
}`

	mutationSetFavoriteFolderAlias = `mutation setFavoriteFolderAlias($fullPath: String!, $alias: String!) {
  setFavoriteFolderAlias(fullPath: $fullPath, alias: $alias) {
    rootPath
    fullPath
    alias
  }
}`

	mutationPlayAudio = `mutation playAudio($path: String!) {
  playAudio(path: $path) {
    ` + playlistAudioFields + `
  }
}`

	mutationUpdateAudioPlayMode = `mutation updateAudioPlayMode($mode: MediaPlayMode!) {
  updateAudioPlayMode(mode: $mode)
}`

	mutationDeletePlaylistAudio = `mutation deletePlaylistAudio($path: String!) {
  deletePlaylistAudio(path: $path)
}`

	mutationAddPlaylistAudios = `mutation addPlaylistAudios($query: String!) {
  addPlaylistAudios(query: $query)
}`

	mutationClearAudioPlaylist = `mutation {
  clearAudioPlaylist
}`

	mutationReorderPlaylistAudios = `mutation reorderPlaylistAudios($paths: [String!]!) {
  reorderPlaylistAudios(paths: $paths)
}`

	mutationDeleteMediaItems = `mutation deleteMediaItems($type: DataType!, $query: String!) {
  deleteMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	mutationTrashMediaItems = `mutation trashMediaItems($type: DataType!, $query: String!) {
  trashMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	mutationRestoreMediaItems = `mutation restoreMediaItems($type: DataType!, $query: String!) {
  restoreMediaItems(type: $type, query: $query) {
    type
    query
  }
}`

	mutationCreateTag = `mutation createTag($type: DataType!, $name: String!) {
  createTag(type: $type, name: $name) {
    ` + tagFields + `
  }
}`

	mutationUpdateTag = `mutation updateTag($id: ID!, $name: String!) {
  updateTag(id: $id, name: $name) {
    ` + tagFields + `
  }
}`

	mutationDeleteTag = `mutation deleteTag($id: ID!) {
  deleteTag(id: $id)
}`

	mutationAddToTags = `mutation addToTags($type: DataType!, $tagIds: [ID!]!, $query: String!) {
  addToTags(type: $type, tagIds: $tagIds, query: $query)
}`

	mutationRemoveFromTags = `mutation removeFromTags($type: DataType!, $tagIds: [ID!]!, $query: String!) {
  removeFromTags(type: $type, tagIds: $tagIds, query: $query)
}`

	mutationUpdateTagRelations = `mutation updateTagRelations($type: DataType!, $item: TagRelationStub!, $addTagIds: [ID!]!, $removeTagIds: [ID!]!) {
  updateTagRelations(type: $type, item: $item, addTagIds: $addTagIds, removeTagIds: $removeTagIds)
}`

	mutationSaveNote = `mutation saveNote($id: ID!, $input: NoteInput!) {
  saveNote(id: $id, input: $input) {
    ` + noteFields + `
  }
}`

	mutationDeleteNotes = `mutation deleteNotes($query: String!) {
  deleteNotes(query: $query)
}`

	mutationTrashNotes = `mutation trashNotes($query: String!) {
  trashNotes(query: $query)
}`

	mutationRestoreNotes = `mutation restoreNotes($query: String!) {
  restoreNotes(query: $query)
}`

	mutationExportNotes = `mutation exportNotes($query: String!) {
  exportNotes(query: $query)
}`

	mutationCreateFeed = `mutation createFeed($url: String!, $fetchContent: Boolean!) {
  createFeed(url: $url, fetchContent: $fetchContent) {
    ` + feedFields + `
  }
}`

	mutationUpdateFeed = `mutation updateFeed($id: ID!, $name: String!, $fetchContent: Boolean!) {
  updateFeed(id: $id, name: $name, fetchContent: $fetchContent) {
    ` + feedFields + `
  }
}`

	mutationDeleteFeed = `mutation deleteFeed($id: ID!) {
  deleteFeed(id: $id)
}`

	mutationSyncFeeds = `mutation syncFeeds($id: ID) {
  syncFeeds(id: $id)
}`

	mutationSyncFeedContent = `mutation syncFeedContent($id: ID!) {
  syncFeedContent(id: $id) {
    ` + feedEntryFields + `
    feed {
      ` + feedFields + `
    }
  }
}`

	mutationImportFeeds = `mutation importFeeds($content: String!) {
  importFeeds(content: $content)
}`

	mutationExportFeeds = `mutation {
  exportFeeds
}`

	mutationDeleteFeedEntries = `mutation deleteFeedEntries($query: String!) {
  deleteFeedEntries(query: $query)
}`

	mutationSaveFeedEntriesToNotes = `mutation saveFeedEntriesToNotes($query: String!) {
  saveFeedEntriesToNotes(query: $query)
}`

	mutationSendSMS = `mutation sendSms($number: String!, $body: String!) {
  sendSms(number: $number, body: $body)
}`

	mutationSendMMS = `mutation sendMms($number: String!, $body: String!, $attachmentPaths: [String!]!, $threadId: String!) {
  sendMms(number: $number, body: $body, attachmentPaths: $attachmentPaths, threadId: $threadId)
}`

	mutationCall = `mutation call($number: String!) {
  call(number: $number)
}`

	mutationDeleteCalls = `mutation deleteCalls($query: String!) {
  deleteCalls(query: $query)
}`

	mutationDeleteContacts = `mutation deleteContacts($query: String!) {
  deleteContacts(query: $query)
}`

	mutationSetClip = `mutation setClip($text: String!) {
  setClip(text: $text)
}`

	mutationInstallPackage = `mutation installPackage($path: String!) {
  installPackage(path: $path) {
    packageName
    updatedAt
    isNew
  }
}`

	mutationUninstallPackages = `mutation uninstallPackages($ids: [ID!]!) {
  uninstallPackages(ids: $ids)
}`

	mutationCancelNotifications = `mutation cancelNotifications($ids: [ID!]!) {
  cancelNotifications(ids: $ids)
}`

	mutationReplyNotification = `mutation replyNotification($id: ID!, $actionIndex: Int!, $text: String!) {
  replyNotification(id: $id, actionIndex: $actionIndex, text: $text)
}`

	mutationAddBookmarks = `mutation addBookmarks($urls: [String!]!, $groupId: String!) {
  addBookmarks(urls: $urls, groupId: $groupId) {
    ` + bookmarkFields + `
  }
}`

	mutationUpdateBookmark = `mutation updateBookmark($id: ID!, $input: BookmarkInput!) {
  updateBookmark(id: $id, input: $input) {
    ` + bookmarkFields + `
  }
}`

	mutationDeleteBookmarks = `mutation deleteBookmarks($ids: [ID!]!) {
  deleteBookmarks(ids: $ids)
}`

	mutationRecordBookmarkClick = `mutation recordBookmarkClick($id: ID!) {
  recordBookmarkClick(id: $id)
}`

	mutationCreateBookmarkGroup = `mutation createBookmarkGroup($name: String!) {
  createBookmarkGroup(name: $name) {
    ` + bookmarkGroupFields + `
  }
}`

	mutationUpdateBookmarkGroup = `mutation updateBookmarkGroup($id: ID!, $name: String!, $collapsed: Boolean!, $sortOrder: Int!) {
  updateBookmarkGroup(id: $id, name: $name, collapsed: $collapsed, sortOrder: $sortOrder) {
    ` + bookmarkGroupFields + `
  }
}`

	mutationDeleteBookmarkGroup = `mutation deleteBookmarkGroup($id: ID!) {
  deleteBookmarkGroup(id: $id)
}`

	mutationStartScreenMirror = `mutation startScreenMirror($audio: Boolean!) {
  startScreenMirror(audio: $audio)
}`

	mutationStopScreenMirror = `mutation {
  stopScreenMirror
}`

	mutationRequestScreenMirrorAudio = `mutation {
  requestScreenMirrorAudio
}`

	mutationUpdateScreenMirrorQuality = `mutation updateScreenMirrorQuality($mode: ScreenMirrorMode!) {
  updateScreenMirrorQuality(mode: $mode)
}`

	mutationSendScreenMirrorControl = `mutation sendScreenMirrorControl($input: ScreenMirrorControlInput!) {
  sendScreenMirrorControl(input: $input)
}`

	mutationSendWebRTCSignaling = `mutation sendWebRtcSignaling($payload: WebRtcSignalingMessage!) {
  sendWebRtcSignaling(payload: $payload)
}`

	mutationStartPomodoro = `mutation startPomodoro($timeLeft: Int!) {
  startPomodoro(timeLeft: $timeLeft)
}`

	mutationStopPomodoro = `mutation {
  stopPomodoro
}`

	mutationPausePomodoro = `mutation {
  pausePomodoro
}`

	mutationEnableImageSearch = `mutation {
  enableImageSearch
}`

	mutationDisableImageSearch = `mutation {
  disableImageSearch
}`

	mutationCancelImageDownload = `mutation {
  cancelImageDownload
}`

	mutationStartImageIndex = `mutation startImageIndex($force: Boolean) {
  startImageIndex(force: $force)
}`

	mutationCancelImageIndex = `mutation {
  cancelImageIndex
}`

	mutationSetTempValue = `mutation setTempValue($key: String!, $value: String!) {
  setTempValue(key: $key, value: $value) {
    key
    value
  }
}`

	mutationRelaunchApp = `mutation {
  relaunchApp
}`
)

var _ = [...]string{
	mutationSendChatItem,
	mutationDeleteChatItem,
	mutationCreateChatChannel,
	mutationUpdateChatChannel,
	mutationDeleteChatChannel,
	mutationLeaveChatChannel,
	mutationAddChatChannelMember,
	mutationRemoveChatChannelMember,
	mutationAcceptChatChannelInvite,
	mutationDeclineChatChannelInvite,
	mutationCreateDir,
	mutationWriteTextFile,
	mutationRenameFile,
	mutationCopyFile,
	mutationMoveFile,
	mutationDeleteFiles,
	mutationMergeChunks,
	mutationAddFavoriteFolder,
	mutationRemoveFavoriteFolder,
	mutationSetFavoriteFolderAlias,
	mutationPlayAudio,
	mutationUpdateAudioPlayMode,
	mutationDeletePlaylistAudio,
	mutationAddPlaylistAudios,
	mutationClearAudioPlaylist,
	mutationReorderPlaylistAudios,
	mutationDeleteMediaItems,
	mutationTrashMediaItems,
	mutationRestoreMediaItems,
	mutationCreateTag,
	mutationUpdateTag,
	mutationDeleteTag,
	mutationAddToTags,
	mutationRemoveFromTags,
	mutationUpdateTagRelations,
	mutationSaveNote,
	mutationDeleteNotes,
	mutationTrashNotes,
	mutationRestoreNotes,
	mutationExportNotes,
	mutationCreateFeed,
	mutationUpdateFeed,
	mutationDeleteFeed,
	mutationSyncFeeds,
	mutationSyncFeedContent,
	mutationImportFeeds,
	mutationExportFeeds,
	mutationDeleteFeedEntries,
	mutationSaveFeedEntriesToNotes,
	mutationSendSMS,
	mutationSendMMS,
	mutationCall,
	mutationDeleteCalls,
	mutationDeleteContacts,
	mutationSetClip,
	mutationInstallPackage,
	mutationUninstallPackages,
	mutationCancelNotifications,
	mutationReplyNotification,
	mutationAddBookmarks,
	mutationUpdateBookmark,
	mutationDeleteBookmarks,
	mutationRecordBookmarkClick,
	mutationCreateBookmarkGroup,
	mutationUpdateBookmarkGroup,
	mutationDeleteBookmarkGroup,
	mutationStartScreenMirror,
	mutationStopScreenMirror,
	mutationRequestScreenMirrorAudio,
	mutationUpdateScreenMirrorQuality,
	mutationSendScreenMirrorControl,
	mutationSendWebRTCSignaling,
	mutationStartPomodoro,
	mutationStopPomodoro,
	mutationPausePomodoro,
	mutationEnableImageSearch,
	mutationDisableImageSearch,
	mutationCancelImageDownload,
	mutationStartImageIndex,
	mutationCancelImageIndex,
	mutationSetTempValue,
	mutationRelaunchApp,
}
