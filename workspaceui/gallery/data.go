package main

// message is a row of the gallery's fixture mail. No name or address here is
// a real one.
type message struct {
	from, address, subject, preview, when string
	unread, flagged, attachment, reminder bool
	thread                                int
}

var inbox = []message{
	{from: "David King", address: "david.king@example.com", subject: "Invitation: Release planning", preview: "David King has invited you to Release planning. Thursday, 10:00 to 11:00, Room 4.", when: "7:15 PM", unread: true},
	{from: "David King", address: "david.king@example.com", subject: "Re: Aurora launch checklist", preview: "The load test finished overnight: 50,000 accounts, p95 sync latency 1.8 s, no errors. Graphs are on the dashboard.", when: "8:41 AM", unread: true, thread: 3},
	{from: "Peter Brown", address: "peter.brown@example.com", subject: "Standup moved to 10:30", preview: "Morning! Today's standup moves to 10:30 so we can join the all-hands. Same room, same link. Peter", when: "7:58 AM", unread: true},
	{from: "Skyline Travel", address: "trips@skyline-travel.example", subject: "Your trip to Lisbon: itinerary", preview: "Your trip to Lisbon, 16-20 October. Booking LX4K2Q. Outbound: Boston (BOS) to Lisbon (LIS), 16 October, 19:05.", when: "6:30 AM", reminder: true},
	{from: "Iris Young", address: "iris@example.com", subject: "Dinner Saturday?", preview: "We're thinking of trying the new Portuguese place on Saturday around 7. Want to come? Bring Sam if he's back.", when: "Yesterday", thread: 2},
	{from: "Lena Price", address: "lena.price@example.com", subject: "Design review notes: onboarding flow", preview: "Ann, Notes from today's review of the onboarding flow: the account step should come before the theme picker.", when: "Yesterday", flagged: true},
	{from: "Omar Ward", address: "omar.ward@example.com", subject: "Q4 roadmap review (slides attached)", preview: "Hi team, Slides for Friday's Q4 roadmap review are attached. Please add questions as comments before then.", when: "Yesterday", attachment: true},
	{from: "The Weekly Frost", address: "letters@weeklyfrost.example", subject: "Issue 42: Gardens that thrive in the cold", preview: "This week: winter greens, cold frames built from old windows, and a reader's quince that fruited at last.", when: "Monday"},
}

// mailbox is a row of the navigation.
type mailbox struct {
	id, name, role string
	count          int
	unread         bool
}

var favorites = []mailbox{
	{"all-inboxes", "All Inboxes", "inbox", 3, true},
	{"flagged", "Flagged", "flag", 2, false},
	{"all-drafts", "All Drafts", "drafts", 1, false},
	{"all-sent", "All Sent", "sent", 0, false},
}

var account = []mailbox{
	{"inbox", "Inbox", "inbox", 3, true},
	{"drafts", "Drafts", "drafts", 1, false},
	{"sent", "Sent", "sent", 0, false},
	{"junk", "Junk", "junk", 12, false},
	{"trash", "Trash", "trash", 0, false},
	{"archive", "Archive", "archive", 0, false},
	{"receipts", "Receipts", "folder", 0, false},
}
