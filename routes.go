package main

import (
	"net/http"
)

func (s *server) routes() {
	// TOOD: Remove media serving Serve media files from disk
	mediaDir := GetFileManager().GetMediaDir()
	s.router.PathPrefix("/media/").Handler(http.StripPrefix("/media/", http.FileServer(http.Dir(mediaDir))))

	s.router.Handle("/health", s.GetHealth()).Methods("GET")

	s.router.Handle("/session/connect", s.Connect()).Methods("POST")
	s.router.Handle("/session/disconnect", s.Disconnect()).Methods("POST")
	s.router.Handle("/session/logout", s.Logout()).Methods("POST")
	s.router.Handle("/session/status", s.GetStatus()).Methods("GET")
	s.router.Handle("/session/qr", s.GetQR()).Methods("GET")
	s.router.Handle("/session/pairphone", s.PairPhone()).Methods("POST")
	s.router.Handle("/session/history", s.RequestHistorySync()).Methods("GET")

	s.router.Handle("/session/history", s.SetHistory()).Methods("POST")

	s.router.Handle("/chat/send/text", s.SendMessage()).Methods("POST")
	s.router.Handle("/chat/delete", s.DeleteMessage()).Methods("POST")
	s.router.Handle("/chat/send/image", s.SendImage()).Methods("POST")
	s.router.Handle("/chat/send/audio", s.SendAudio()).Methods("POST")
	s.router.Handle("/chat/send/document", s.SendDocument()).Methods("POST")
	s.router.Handle("/chat/send/video", s.SendVideo()).Methods("POST")
	s.router.Handle("/chat/send/sticker", s.SendSticker()).Methods("POST")
	s.router.Handle("/chat/send/location", s.SendLocation()).Methods("POST")
	s.router.Handle("/chat/send/contact", s.SendContact()).Methods("POST")
	s.router.Handle("/chat/react", s.React()).Methods("POST")
	s.router.Handle("/chat/send/buttons", s.SendButtons()).Methods("POST")
	s.router.Handle("/chat/send/list", s.SendList()).Methods("POST")
	s.router.Handle("/chat/send/poll", s.SendPoll()).Methods("POST")
	s.router.Handle("/chat/send/edit", s.SendEditMessage()).Methods("POST")
	s.router.Handle("/chat/request-unavailable-message", s.RequestUnavailableMessage()).Methods("POST")
	s.router.Handle("/chat/archive", s.ArchiveChat()).Methods("POST")

	s.router.Handle("/status/set/text", s.SetStatusMessage()).Methods("POST")

	s.router.Handle("/call/reject", s.RejectCall()).Methods("POST")

	s.router.Handle("/user/presence", s.SendPresence()).Methods("POST")
	s.router.Handle("/user/info", s.GetUser()).Methods("POST")
	s.router.Handle("/user/check", s.CheckUser()).Methods("POST")
	s.router.Handle("/user/avatar", s.GetAvatar()).Methods("POST")
	s.router.Handle("/user/contacts", s.GetContacts()).Methods("GET")
	s.router.Handle("/user/lid/{jid}", s.GetUserLID()).Methods("GET")

	s.router.Handle("/chat/presence", s.ChatPresence()).Methods("POST")
	s.router.Handle("/chat/markread", s.MarkRead()).Methods("POST")
	s.router.Handle("/chat/downloadimage", s.DownloadImage()).Methods("POST")
	s.router.Handle("/chat/downloadvideo", s.DownloadVideo()).Methods("POST")
	s.router.Handle("/chat/downloadaudio", s.DownloadAudio()).Methods("POST")
	s.router.Handle("/chat/downloaddocument", s.DownloadDocument()).Methods("POST")
	s.router.Handle("/chat/downloadsticker", s.DownloadSticker()).Methods("POST")

	s.router.Handle("/group/create", s.CreateGroup()).Methods("POST")
	s.router.Handle("/group/list", s.ListGroups()).Methods("GET")
	s.router.Handle("/group/info", s.GetGroupInfo()).Methods("GET")
	s.router.Handle("/group/invitelink", s.GetGroupInviteLink()).Methods("GET")
	s.router.Handle("/group/photo", s.SetGroupPhoto()).Methods("POST")
	s.router.Handle("/group/photo/remove", s.RemoveGroupPhoto()).Methods("POST")
	s.router.Handle("/group/leave", s.GroupLeave()).Methods("POST")
	s.router.Handle("/group/name", s.SetGroupName()).Methods("POST")
	s.router.Handle("/group/topic", s.SetGroupTopic()).Methods("POST")
	s.router.Handle("/group/announce", s.SetGroupAnnounce()).Methods("POST")
	s.router.Handle("/group/locked", s.SetGroupLocked()).Methods("POST")
	s.router.Handle("/group/ephemeral", s.SetDisappearingTimer()).Methods("POST")
	s.router.Handle("/group/join", s.GroupJoin()).Methods("POST")
	s.router.Handle("/group/inviteinfo", s.GetGroupInviteInfo()).Methods("POST")
	s.router.Handle("/group/updateparticipants", s.UpdateGroupParticipants()).Methods("POST")

	s.router.Handle("/newsletter/list", s.ListNewsletter()).Methods("GET")
}
