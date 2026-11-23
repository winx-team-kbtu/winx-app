# Team conversations

A Team Mode like now creates one shared conversation for that pair and course.
It does not require a reciprocal like. The choice and chat commit in one PostgreSQL
transaction; retries and simultaneous reciprocal likes do not create duplicates.
Previously saved likes are migrated into conversations.

Team Mode has four destinations: search, chats, team and profile. The chat list
shows peer names, last messages and dates, filtered by course. Incoming likes are
visible to the recipient. The old `/team/shortlist` route redirects to chats.

Conversations reuse the existing messages table and WebSocket transport.
Academic chat metadata/history/decisions use the authenticated profile academic
API. Participants can read their history; outsiders receive 404. Social chat lists
exclude academic rooms. The frontend normalizes the server's nested message
envelope, waits for a server acknowledgment and merges history by message ID.

Use **Пригласить в команду** to invite the peer to your current team. If you
have no team for the course, enter a name to create one and send the invitation.
Incoming invitations appear in the chat list and in the conversation. Only the
recipient can accept or reject an invitation. Accepting adds the recipient to
the roster. Rejecting leaves the conversation open.

Use **Написать** on a teammate's profile to open your conversation for the course.
Leaving a team preserves the conversation. To remove a course from your profile,
leave its team first.

Acceptance checks current course enrollment, membership and available space.
The team row lock serializes accepts and joining by code to enforce the five-person
limit. **Завершить общение** hides the chat from both active lists and closes it
to new messages. Participants can still read the history.

The invitation table is installed by the combined, timestamp-ordered migration
set that `scripts/start-local.ps1` assembles.

Integration tests use `WINX_TEST_DSN` and explicitly require a database named
`winx_conv_qa_*`. Run with `go test -tags integration` in the conversation
repository package and the chat message repository package. Unit tests cover a
stale WebSocket disconnect after access-token rotation.
