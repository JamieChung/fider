package postgres_test

import (
	"testing"
	"time"

	"github.com/getfider/fider/app"
	"github.com/getfider/fider/app/models"
	"github.com/getfider/fider/app/pkg/errors"
	. "github.com/onsi/gomega"
)

func TestIdeaStorage_GetAll(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	now := time.Now()

	trx.Execute("INSERT INTO ideas (title, slug, number, description, created_on, tenant_id, user_id, supporters, status) VALUES ('add twitter integration', 'add-twitter-integration', 1, 'Would be great to see it integrated with twitter', $1, 1, 1, 0, 1)", now)
	trx.Execute("INSERT INTO ideas (title, slug, number, description, created_on, tenant_id, user_id, supporters, status) VALUES ('this is my idea', 'this-is-my-idea', 2, 'no description', $1, 1, 2, 5, 2)", now)

	ideas.SetCurrentTenant(demoTenant)

	dbIdeas, err := ideas.GetAll()
	Expect(err).To(BeNil())
	Expect(dbIdeas).To(HaveLen(2))

	Expect(dbIdeas[0].Title).To(Equal("this is my idea"))
	Expect(dbIdeas[0].Slug).To(Equal("this-is-my-idea"))
	Expect(dbIdeas[0].Number).To(Equal(2))
	Expect(dbIdeas[0].Description).To(Equal("no description"))
	Expect(dbIdeas[0].User.Name).To(Equal("Arya Stark"))
	Expect(dbIdeas[0].TotalSupporters).To(Equal(5))
	Expect(dbIdeas[0].Status).To(Equal(models.IdeaCompleted))

	Expect(dbIdeas[1].Title).To(Equal("add twitter integration"))
	Expect(dbIdeas[1].Slug).To(Equal("add-twitter-integration"))
	Expect(dbIdeas[1].Number).To(Equal(1))
	Expect(dbIdeas[1].Description).To(Equal("Would be great to see it integrated with twitter"))
	Expect(dbIdeas[1].User.Name).To(Equal("Jon Snow"))
	Expect(dbIdeas[1].TotalSupporters).To(Equal(0))
	Expect(dbIdeas[1].Status).To(Equal(models.IdeaStarted))

	dbIdeas, err = ideas.Search("twitter", "trending", []string{})
	Expect(err).To(BeNil())
	Expect(dbIdeas).To(HaveLen(1))
	Expect(dbIdeas[0].Slug).To(Equal("add-twitter-integration"))
}

func TestIdeaStorage_AddAndGet(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	dbIdeaById, err := ideas.GetByID(idea.ID)
	Expect(err).To(BeNil())
	Expect(dbIdeaById.ID).To(Equal(idea.ID))
	Expect(dbIdeaById.Number).To(Equal(1))
	Expect(dbIdeaById.ViewerSupported).To(BeFalse())
	Expect(dbIdeaById.TotalSupporters).To(Equal(0))
	Expect(dbIdeaById.Status).To(Equal(models.IdeaOpen))
	Expect(dbIdeaById.Title).To(Equal("My new idea"))
	Expect(dbIdeaById.Description).To(Equal("with this description"))
	Expect(dbIdeaById.User.ID).To(Equal(1))
	Expect(dbIdeaById.User.Name).To(Equal("Jon Snow"))
	Expect(dbIdeaById.User.Email).To(Equal("jon.snow@got.com"))

	dbIdeaBySlug, err := ideas.GetBySlug("my-new-idea")

	Expect(err).To(BeNil())
	Expect(dbIdeaBySlug.ID).To(Equal(idea.ID))
	Expect(dbIdeaBySlug.Number).To(Equal(1))
	Expect(dbIdeaBySlug.ViewerSupported).To(BeFalse())
	Expect(dbIdeaBySlug.TotalSupporters).To(Equal(0))
	Expect(dbIdeaBySlug.Status).To(Equal(models.IdeaOpen))
	Expect(dbIdeaBySlug.Title).To(Equal("My new idea"))
	Expect(dbIdeaBySlug.Description).To(Equal("with this description"))
	Expect(dbIdeaBySlug.User.ID).To(Equal(1))
	Expect(dbIdeaBySlug.User.Name).To(Equal("Jon Snow"))
	Expect(dbIdeaBySlug.User.Email).To(Equal("jon.snow@got.com"))
}

func TestIdeaStorage_GetInvalid(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)

	dbIdea, err := ideas.GetByID(1)
	Expect(errors.Cause(err)).To(Equal(app.ErrNotFound))
	Expect(dbIdea).To(BeNil())
}

func TestIdeaStorage_AddAndReturnComments(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	ideas.SetCurrentUser(jonSnow)
	ideas.AddComment(idea, "Comment #1")
	ideas.SetCurrentUser(aryaStark)
	ideas.AddComment(idea, "Comment #2")

	comments, err := ideas.GetCommentsByIdea(idea)
	Expect(err).To(BeNil())
	Expect(len(comments)).To(Equal(2))

	Expect(comments[0].Content).To(Equal("Comment #1"))
	Expect(comments[0].User.Name).To(Equal("Jon Snow"))
	Expect(comments[1].Content).To(Equal("Comment #2"))
	Expect(comments[1].User.Name).To(Equal("Arya Stark"))
}

func TestIdeaStorage_AddGetUpdateComment(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	commentId, err := ideas.AddComment(idea, "Comment #1")
	Expect(err).To(BeNil())

	comment, err := ideas.GetCommentByID(commentId)
	Expect(err).To(BeNil())
	Expect(comment.ID).To(Equal(commentId))
	Expect(comment.Content).To(Equal("Comment #1"))
	Expect(comment.User.ID).To(Equal(jonSnow.ID))
	Expect(comment.EditedOn).To(BeNil())
	Expect(comment.EditedBy).To(BeNil())

	ideas.SetCurrentUser(aryaStark)
	err = ideas.UpdateComment(commentId, "Comment #1 with edit")
	Expect(err).To(BeNil())

	comment, err = ideas.GetCommentByID(commentId)
	Expect(err).To(BeNil())
	Expect(comment.ID).To(Equal(commentId))
	Expect(comment.Content).To(Equal("Comment #1 with edit"))
	Expect(comment.User.ID).To(Equal(jonSnow.ID))
	Expect(comment.EditedOn).NotTo(BeNil())
	Expect(comment.EditedBy.ID).To(Equal(aryaStark.ID))
}

func TestIdeaStorage_AddAndGet_DifferentTenants(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	demoIdea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	ideas.SetCurrentTenant(avengersTenant)
	ideas.SetCurrentUser(tonyStark)
	avengersIdea, err := ideas.Add("My other idea", "with other description")
	Expect(err).To(BeNil())

	ideas.SetCurrentTenant(demoTenant)
	dbIdea, err := ideas.GetByNumber(1)

	Expect(err).To(BeNil())
	Expect(dbIdea.ID).To(Equal(demoIdea.ID))
	Expect(dbIdea.Number).To(Equal(1))
	Expect(dbIdea.Title).To(Equal("My new idea"))
	Expect(dbIdea.Slug).To(Equal("my-new-idea"))

	ideas.SetCurrentTenant(avengersTenant)
	dbIdea, err = ideas.GetByNumber(1)

	Expect(err).To(BeNil())
	Expect(dbIdea.ID).To(Equal(avengersIdea.ID))
	Expect(dbIdea.Number).To(Equal(1))
	Expect(dbIdea.Title).To(Equal("My other idea"))
	Expect(dbIdea.Slug).To(Equal("my-other-idea"))
}

func TestIdeaStorage_Update(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	idea, err = ideas.Update(idea, "The new comment", "With the new description")
	Expect(err).To(BeNil())
	Expect(idea.Title).To(Equal("The new comment"))
	Expect(idea.Description).To(Equal("With the new description"))
	Expect(idea.Slug).To(Equal("the-new-comment"))
}

func TestIdeaStorage_AddSupporter(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	users.SetCurrentTenant(demoTenant)
	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	err = ideas.AddSupporter(idea, aryaStark)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(1)
	Expect(dbIdea.ViewerSupported).To(BeFalse())
	Expect(dbIdea.TotalSupporters).To(Equal(1))

	err = ideas.AddSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	dbIdea, err = ideas.GetByNumber(1)
	Expect(err).To(BeNil())
	Expect(dbIdea.ViewerSupported).To(BeTrue())
	Expect(dbIdea.TotalSupporters).To(Equal(2))
}

func TestIdeaStorage_AddSupporter_Twice(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")

	err := ideas.AddSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	err = ideas.AddSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(1))
}

func TestIdeaStorage_RemoveSupporter(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")

	err := ideas.AddSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	err = ideas.RemoveSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(0))
}

func TestIdeaStorage_RemoveSupporter_Twice(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")

	err := ideas.AddSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	err = ideas.RemoveSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	err = ideas.RemoveSupporter(idea, jonSnow)
	Expect(err).To(BeNil())

	dbIdea, err := ideas.GetByNumber(1)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(0))
}

func TestIdeaStorage_SetResponse(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")
	err := ideas.SetResponse(idea, "We liked this idea", models.IdeaStarted)

	Expect(err).To(BeNil())

	idea, _ = ideas.GetByID(idea.ID)
	Expect(idea.Response.Text).To(Equal("We liked this idea"))
	Expect(idea.Status).To(Equal(models.IdeaStarted))
	Expect(idea.Response.User.ID).To(Equal(1))
}

func TestIdeaStorage_SetResponse_KeepOpen(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")
	err := ideas.SetResponse(idea, "We liked this idea", models.IdeaOpen)
	Expect(err).To(BeNil())
}

func TestIdeaStorage_SetResponse_ChangeText(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")
	ideas.SetResponse(idea, "We liked this idea", models.IdeaStarted)
	idea, _ = ideas.GetByID(idea.ID)
	respondedOn := idea.Response.RespondedOn

	ideas.SetResponse(idea, "We liked this idea and we'll work on it", models.IdeaStarted)
	idea, _ = ideas.GetByID(idea.ID)
	Expect(idea.Response.RespondedOn).To(Equal(respondedOn))

	ideas.SetResponse(idea, "We finished it", models.IdeaCompleted)
	idea, _ = ideas.GetByID(idea.ID)
	Expect(idea.Response.RespondedOn).Should(BeTemporally(">", respondedOn))
}

func TestIdeaStorage_SetResponse_AsDuplicate(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)

	idea1, _ := ideas.Add("My new idea", "with this description")
	ideas.AddSupporter(idea1, jonSnow)

	ideas.SetCurrentUser(aryaStark)
	idea2, _ := ideas.Add("My other idea", "with similar description")
	ideas.AddSupporter(idea2, aryaStark)

	ideas.SetCurrentUser(jonSnow)
	ideas.MarkAsDuplicate(idea2, idea1)
	idea1, _ = ideas.GetByID(idea1.ID)

	Expect(idea1.TotalSupporters).To(Equal(2))
	Expect(idea1.Status).To(Equal(models.IdeaOpen))
	Expect(idea1.Response).To(BeNil())

	idea2, _ = ideas.GetByID(idea2.ID)

	Expect(idea2.Response.Text).To(Equal(""))
	Expect(idea2.TotalSupporters).To(Equal(1))
	Expect(idea2.Status).To(Equal(models.IdeaDuplicate))
	Expect(idea2.Response.User.ID).To(Equal(1))
	Expect(idea2.Response.Original.Number).To(Equal(idea1.Number))
	Expect(idea2.Response.Original.Title).To(Equal(idea1.Title))
	Expect(idea2.Response.Original.Slug).To(Equal(idea1.Slug))
	Expect(idea2.Response.Original.Status).To(Equal(idea1.Status))
}

func TestIdeaStorage_SetResponse_AsDeleted(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea, err := ideas.Add("My new idea", "with this description")
	Expect(err).To(BeNil())

	ideas.SetResponse(idea, "Spam!", models.IdeaDeleted)

	idea1, err := ideas.GetByNumber(idea.Number)
	Expect(errors.Cause(err)).To(Equal(app.ErrNotFound))
	Expect(idea1).To(BeNil())

	idea2, err := ideas.GetByID(idea.ID)
	Expect(errors.Cause(err)).To(Equal(app.ErrNotFound))
	Expect(idea2).To(BeNil())
}

func TestIdeaStorage_AddSupporter_ClosedIdea(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea, _ := ideas.Add("My new idea", "with this description")
	ideas.SetResponse(idea, "We liked this idea", models.IdeaCompleted)
	ideas.AddSupporter(idea, jonSnow)

	dbIdea, err := ideas.GetByNumber(idea.Number)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(0))
}

func TestIdeaStorage_RemoveSupporter_ClosedIdea(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea, _ := ideas.Add("My new idea", "with this description")
	ideas.AddSupporter(idea, jonSnow)
	ideas.SetResponse(idea, "We liked this idea", models.IdeaCompleted)
	ideas.RemoveSupporter(idea, jonSnow)

	dbIdea, err := ideas.GetByNumber(idea.Number)
	Expect(err).To(BeNil())
	Expect(dbIdea.TotalSupporters).To(Equal(1))
}

func TestIdeaStorage_ListSupportedIdeas(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea1, _ := ideas.Add("My new idea", "with this description")
	idea2, _ := ideas.Add("My other idea", "with better description")
	ideas.AddSupporter(idea1, aryaStark)
	ideas.AddSupporter(idea2, aryaStark)

	ideas.SetCurrentUser(jonSnow)
	Expect(ideas.SupportedBy()).To(Equal([]int{}))

	ideas.SetCurrentUser(aryaStark)
	Expect(ideas.SupportedBy()).To(Equal([]int{idea1.ID, idea2.ID}))
}

func TestIdeaStorage_WithTags(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(aryaStark)
	tags.SetCurrentTenant(demoTenant)
	tags.SetCurrentUser(jonSnow)

	idea, _ := ideas.Add("My new idea", "with this description")
	bug, _ := tags.Add("Bug", "FF0000", true)
	featureRequest, _ := tags.Add("Feature Request", "00FF00", false)

	tags.AssignTag(bug, idea)
	tags.AssignTag(featureRequest, idea)

	idea, _ = ideas.GetByID(idea.ID)
	Expect(len(idea.Tags)).To(Equal(1))
	Expect(idea.Tags[0]).To(Equal(bug.Slug))

	ideas.SetCurrentUser(jonSnow)
	idea, _ = ideas.GetByID(idea.ID)
	Expect(len(idea.Tags)).To(Equal(2))
	Expect(idea.Tags[0]).To(Equal(bug.Slug))
	Expect(idea.Tags[1]).To(Equal(featureRequest.Slug))
}

func TestIdeaStorage_IsReferenced(t *testing.T) {
	SetupDatabaseTest(t)
	defer TeardownDatabaseTest()

	ideas.SetCurrentTenant(demoTenant)
	ideas.SetCurrentUser(jonSnow)
	idea1, _ := ideas.Add("My first idea", "with this description")
	idea2, _ := ideas.Add("My second idea", "with this description")
	idea3, _ := ideas.Add("My third idea", "with this description")

	ideas.MarkAsDuplicate(idea2, idea3)
	ideas.MarkAsDuplicate(idea3, idea1)

	Expect(ideas.IsReferenced(idea1)).To(BeTrue())
	Expect(ideas.IsReferenced(idea2)).To(BeFalse())
	Expect(ideas.IsReferenced(idea3)).To(BeTrue())
}
