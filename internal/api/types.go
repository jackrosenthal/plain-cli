package api

import (
	"encoding/json"
	"strings"
)

type Tag struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type File struct {
	Path      string `json:"path"`
	IsDir     bool   `json:"isDir"`
	Size      int    `json:"size"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Children  int    `json:"children"`
	MediaID   string `json:"mediaId"`
}

type Image struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	Size      int    `json:"size"`
	BucketID  string `json:"bucketId"`
	TakenAt   string `json:"takenAt"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Tags      []Tag  `json:"tags"`
}

type Video struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Path      string `json:"path"`
	Duration  int    `json:"duration"`
	Size      int    `json:"size"`
	BucketID  string `json:"bucketId"`
	TakenAt   string `json:"takenAt"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Tags      []Tag  `json:"tags"`
}

type Audio struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Artist      string `json:"artist"`
	Path        string `json:"path"`
	Duration    int    `json:"duration"`
	Size        int    `json:"size"`
	BucketID    string `json:"bucketId"`
	AlbumFileID string `json:"albumFileId"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Tags        []Tag  `json:"tags"`
}

type Note struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	DeletedAt string `json:"deletedAt"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Tags      []Tag  `json:"tags"`
}

type Feed struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url"`
	FetchContent bool   `json:"fetchContent"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type FeedEntry struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	Image       string `json:"image"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Content     string `json:"content"`
	FeedID      string `json:"feedId"`
	RawID       string `json:"rawId"`
	PublishedAt string `json:"publishedAt"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
	Tags        []Tag  `json:"tags"`
}

type MessageAttachment struct {
	Path        string `json:"path"`
	ContentType string `json:"contentType"`
	Name        string `json:"name"`
}

type Message struct {
	ID             string              `json:"id"`
	Body           string              `json:"body"`
	Address        string              `json:"address"`
	ServiceCenter  string              `json:"serviceCenter"`
	Date           string              `json:"date"`
	Type           int                 `json:"type"`
	ThreadID       string              `json:"threadId"`
	SubscriptionID int                 `json:"subscriptionId"`
	IsMMS          bool                `json:"isMms"`
	Attachments    []MessageAttachment `json:"attachments"`
	Tags           []Tag               `json:"tags"`
}

type MessageConversation struct {
	ID           string `json:"id"`
	Address      string `json:"address"`
	Snippet      string `json:"snippet"`
	Date         string `json:"date"`
	MessageCount int    `json:"messageCount"`
	Read         bool   `json:"read"`
}

type ContentItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

type PhoneNumber struct {
	Label            string `json:"label"`
	Value            string `json:"value"`
	Type             string `json:"type"`
	NormalizedNumber string `json:"normalizedNumber"`
}

type Contact struct {
	ID           string        `json:"id"`
	Prefix       string        `json:"prefix"`
	Suffix       string        `json:"suffix"`
	FirstName    string        `json:"firstName"`
	MiddleName   string        `json:"middleName"`
	LastName     string        `json:"lastName"`
	UpdatedAt    string        `json:"updatedAt"`
	Notes        string        `json:"notes"`
	Source       string        `json:"source"`
	ThumbnailID  string        `json:"thumbnailId"`
	Starred      bool          `json:"starred"`
	PhoneNumbers []PhoneNumber `json:"phoneNumbers"`
	Addresses    []ContentItem `json:"addresses"`
	Emails       []ContentItem `json:"emails"`
	Websites     []ContentItem `json:"websites"`
	Events       []ContentItem `json:"events"`
	IMS          []ContentItem `json:"ims"`
	Tags         []Tag         `json:"tags"`
}

type Geo struct {
	ISP      int    `json:"isp"`
	City     string `json:"city"`
	Province string `json:"province"`
}

type Call struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Number    string `json:"number"`
	Duration  int    `json:"duration"`
	AccountID string `json:"accountId"`
	StartedAt string `json:"startedAt"`
	PhotoID   string `json:"photoId"`
	Type      int    `json:"type"`
	Geo       Geo    `json:"geo"`
	Tags      []Tag  `json:"tags"`
}

type PackageCert struct {
	Issuer       string `json:"issuer"`
	Subject      string `json:"subject"`
	SerialNumber string `json:"serialNumber"`
	ValidFrom    string `json:"validFrom"`
	ValidTo      string `json:"validTo"`
}

type Package struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Type        string        `json:"type"`
	Version     string        `json:"version"`
	Path        string        `json:"path"`
	Size        int           `json:"size"`
	Certs       []PackageCert `json:"certs"`
	InstalledAt string        `json:"installedAt"`
	UpdatedAt   string        `json:"updatedAt"`
}

type Notification struct {
	ID           string   `json:"id"`
	OnlyOnce     bool     `json:"onlyOnce"`
	IsClearable  bool     `json:"isClearable"`
	AppID        string   `json:"appId"`
	AppName      string   `json:"appName"`
	Time         string   `json:"time"`
	Silent       bool     `json:"silent"`
	Title        string   `json:"title"`
	Body         string   `json:"body"`
	Actions      []string `json:"actions"`
	ReplyActions []string `json:"replyActions"`
}

type Bookmark struct {
	ID            string `json:"id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
	FaviconPath   string `json:"faviconPath"`
	GroupID       string `json:"groupId"`
	Pinned        bool   `json:"pinned"`
	ClickCount    int    `json:"clickCount"`
	LastClickedAt string `json:"lastClickedAt"`
	SortOrder     int    `json:"sortOrder"`
	CreatedAt     string `json:"createdAt"`
	UpdatedAt     string `json:"updatedAt"`
}

type BookmarkGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Collapsed bool   `json:"collapsed"`
	SortOrder int    `json:"sortOrder"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type MessageImages struct {
	IDs []string `json:"ids"`
}

type MessageFiles struct {
	IDs []string `json:"ids"`
}

type MessageText struct {
	IDs []string `json:"ids"`
}

type ChatItem struct {
	ID        string          `json:"id"`
	FromID    string          `json:"fromId"`
	ToID      string          `json:"toId"`
	ChannelID string          `json:"channelId"`
	CreatedAt string          `json:"createdAt"`
	Content   string          `json:"content"`
	Data      json.RawMessage `json:"data"`
}

type ChatChannelMember struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ChatChannel struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Owner     string              `json:"owner"`
	Members   []ChatChannelMember `json:"members"`
	Version   int                 `json:"version"`
	Status    string              `json:"status"`
	CreatedAt string              `json:"createdAt"`
	UpdatedAt string              `json:"updatedAt"`
}

type Peer struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IP         string `json:"ip"`
	Status     string `json:"status"`
	Port       int    `json:"port"`
	DeviceType string `json:"deviceType"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type Mount struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Path       string `json:"path"`
	MountPoint string `json:"mountPoint"`
	FSType     string `json:"fsType"`
	TotalBytes int    `json:"totalBytes"`
	UsedBytes  int    `json:"usedBytes"`
	FreeBytes  int    `json:"freeBytes"`
	Remote     bool   `json:"remote"`
	Alias      string `json:"alias"`
	DriveType  string `json:"driveType"`
	DiskID     string `json:"diskID"`
}

type FavoriteFolder struct {
	RootPath string `json:"rootPath"`
	FullPath string `json:"fullPath"`
	Alias    string `json:"alias"`
}

type App struct {
	USBConnected        bool             `json:"usbConnected"`
	URLToken            string           `json:"urlToken"`
	HTTPPort            int              `json:"httpPort"`
	HTTPSPort           int              `json:"httpsPort"`
	AppDir              string           `json:"appDir"`
	DeviceName          string           `json:"deviceName"`
	Battery             int              `json:"battery"`
	AppVersion          string           `json:"appVersion"`
	OSVersion           string           `json:"osVersion"`
	Channel             string           `json:"channel"`
	Permissions         []string         `json:"permissions"`
	Audios              []PlaylistAudio  `json:"audios"`
	AudioCurrent        string           `json:"audioCurrent"`
	AudioMode           string           `json:"audioMode"`
	SDCardPath          string           `json:"sdcardPath"`
	USBDiskPaths        []string         `json:"usbDiskPaths"`
	InternalStoragePath string           `json:"internalStoragePath"`
	DownloadsDir        string           `json:"downloadsDir"`
	DeveloperMode       bool             `json:"developerMode"`
	FavoriteFolders     []FavoriteFolder `json:"favoriteFolders"`
}

type PlaylistAudio struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Path     string `json:"path"`
	Duration int    `json:"duration"`
}

type DevicePhoneNumber struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Number string `json:"number"`
}

type DeviceInfo struct {
	DeviceName          string              `json:"deviceName"`
	ReleaseBuildVersion string              `json:"releaseBuildVersion"`
	VersionCodeName     string              `json:"versionCodeName"`
	Manufacturer        string              `json:"manufacturer"`
	SecurityPatch       string              `json:"securityPatch"`
	Bootloader          string              `json:"bootloader"`
	DeviceID            string              `json:"deviceId"`
	Model               string              `json:"model"`
	Product             string              `json:"product"`
	Fingerprint         string              `json:"fingerprint"`
	Hardware            string              `json:"hardware"`
	RadioVersion        string              `json:"radioVersion"`
	Device              string              `json:"device"`
	Board               string              `json:"board"`
	DisplayVersion      string              `json:"displayVersion"`
	BuildBrand          string              `json:"buildBrand"`
	BuildHost           string              `json:"buildHost"`
	BuildTime           string              `json:"buildTime"`
	Uptime              string              `json:"uptime"`
	BuildUser           string              `json:"buildUser"`
	Serial              string              `json:"serial"`
	OSVersion           string              `json:"osVersion"`
	Language            string              `json:"language"`
	SDKVersion          string              `json:"sdkVersion"`
	JavaVMVersion       string              `json:"javaVmVersion"`
	KernelVersion       string              `json:"kernelVersion"`
	GLEsVersion         string              `json:"glEsVersion"`
	ScreenDensity       int                 `json:"screenDensity"`
	ScreenHeight        int                 `json:"screenHeight"`
	ScreenWidth         int                 `json:"screenWidth"`
	PhoneNumbers        []DevicePhoneNumber `json:"phoneNumbers"`
}

type Battery struct {
	Level       int     `json:"level"`
	Voltage     int     `json:"voltage"`
	Health      string  `json:"health"`
	Plugged     bool    `json:"plugged"`
	Temperature float64 `json:"temperature"`
	Status      string  `json:"status"`
	Technology  string  `json:"technology"`
	Capacity    int     `json:"capacity"`
}

type PomodoroToday struct {
	Date           string `json:"date"`
	CompletedCount int    `json:"completedCount"`
	CurrentRound   int    `json:"currentRound"`
	TimeLeft       int    `json:"timeLeft"`
	TotalTime      int    `json:"totalTime"`
	IsRunning      bool   `json:"isRunning"`
	IsPause        bool   `json:"isPause"`
	State          string `json:"state"`
}

type PomodoroSettings struct {
	WorkDuration             int  `json:"workDuration"`
	ShortBreakDuration       int  `json:"shortBreakDuration"`
	LongBreakDuration        int  `json:"longBreakDuration"`
	PomodorosBeforeLongBreak int  `json:"pomodorosBeforeLongBreak"`
	ShowNotification         bool `json:"showNotification"`
	PlaySoundOnComplete      bool `json:"playSoundOnComplete"`
}

type ImageSearchStatus struct {
	Status           string  `json:"status"`
	DownloadProgress float64 `json:"downloadProgress"`
	ErrorMessage     string  `json:"errorMessage"`
	ModelSize        int     `json:"modelSize"`
	ModelDir         string  `json:"modelDir"`
	IsIndexing       bool    `json:"isIndexing"`
	TotalImages      int     `json:"totalImages"`
	IndexedImages    int     `json:"indexedImages"`
}

type ScreenMirrorQuality struct {
	Mode       ScreenMirrorMode `json:"mode"`
	Resolution string           `json:"resolution"`
}

type DataType string

const (
	DataTypeImage     DataType = "image"
	DataTypeVideo     DataType = "video"
	DataTypeAudio     DataType = "audio"
	DataTypeNote      DataType = "note"
	DataTypeFeedEntry DataType = "feed-entry"
	DataTypeCall      DataType = "call"
	DataTypeContact   DataType = "contact"
	DataTypeMessage   DataType = "message"
	DataTypeBookmark  DataType = "bookmark"
)

func (d DataType) ToGraphQL() string {
	return toGraphQLEnum(string(d))
}

type FileSortBy string

const (
	FileSortByName     FileSortBy = "name"
	FileSortByNameDesc FileSortBy = "name-desc"
	FileSortBySize     FileSortBy = "size"
	FileSortBySizeDesc FileSortBy = "size-desc"
	FileSortByDate     FileSortBy = "date"
	FileSortByDateDesc FileSortBy = "date-desc"
)

func (f FileSortBy) ToGraphQL() string {
	return toGraphQLEnum(string(f))
}

type MediaPlayMode string

const (
	MediaPlayModeOrder     MediaPlayMode = "order"
	MediaPlayModeShuffle   MediaPlayMode = "shuffle"
	MediaPlayModeRepeat    MediaPlayMode = "repeat"
	MediaPlayModeRepeatOne MediaPlayMode = "repeat-one"
)

func (m MediaPlayMode) ToGraphQL() string {
	return toGraphQLEnum(string(m))
}

type ScreenMirrorMode string

const (
	ScreenMirrorModeAuto   ScreenMirrorMode = "auto"
	ScreenMirrorModeHD     ScreenMirrorMode = "hd"
	ScreenMirrorModeSmooth ScreenMirrorMode = "smooth"
)

func (s ScreenMirrorMode) ToGraphQL() string {
	return toGraphQLEnum(string(s))
}

func toGraphQLEnum(value string) string {
	return strings.ToUpper(strings.ReplaceAll(value, "-", "_"))
}
