package api

import (
	"time"

	"github.com/seccom/kpass/src/auth"
	"github.com/seccom/kpass/src/schema"
	"github.com/seccom/kpass/src/util"
	"github.com/teambition/gear"
)

// Share is API oject for shares
//
// @Name Share
// @Description Share API
// @Accepts json
// @Produces json
type Share struct {
	CommonAPI
}

// Init ...
func (a *Share) Init(api CommonAPI) *Share {
	a.CommonAPI = api
	return a
}

type tplShareCreate struct {
	Name   string `json:"name" swaggo:"true,share name,Github"`
	Pass   string `json:"pass" swaggo:"false,team password hashed by sha256,15e2536def2490c115759ceabf012872fddbd7887fbe67e5074d1e66148d5d00"`
	UserID string `json:"userID" swaggo:"true,user id share to,jeo"`
	Expire int    `json:"expire" swaggo:"true,expire time in seconds,36000"`
}

func (t *tplShareCreate) Validate() error {
	if t.Name == "" {
		return gear.ErrBadRequest.WithMsg("invalid share name")
	}
	if !util.IsHashString(t.Pass) {
		return gear.ErrBadRequest.WithMsg("invalid share pass, pass should be hashed by sha256")
	}
	if t.UserID == "" {
		return gear.ErrBadRequest.WithMsg("invalid user ID to share")
	}
	if t.Expire < 10 {
		return gear.ErrBadRequest.WithMsg("invalid share expire time")
	}
	return nil
}

// Create ...
//
// @Title Create
// @Summary Create a share of the entry
// @Description all team members can create share
// @Param Authorization header string true "access_token"
// @Param entryID path string true "entry ID"
// @Param body body tplShareCreate true "share body"
// @Success 200 schema.ShareResult
// @Failure 400 string
// @Failure 401 string
// @Router POST /api/entries/{entryID}/shares
func (a *Share) Create(ctx *gear.Context) (err error) {
	EntryID, err := util.ParseOID(ctx.Param("entryID"))
	if err != nil {
		return gear.ErrBadRequest.From(err)
	}

	body := new(tplShareCreate)
	if err := ctx.ParseBody(body); err != nil {
		return gear.ErrBadRequest.From(err)
	}
	if err = a.models.User.CheckID(body.UserID); err != nil {
		return gear.ErrBadRequest.From(err)
	}

	entry, err := a.models.Entry.Find(EntryID, false)
	if err != nil {
		return gear.ErrNotFound.From(err)
	}
	key, err := auth.KeyFromCtx(ctx)
	if err != nil {
		return gear.ErrUnauthorized.From(err)
	}
	userID, _ := auth.UserIDFromCtx(ctx)
	if err = a.models.Team.CheckMember(entry.TeamID, userID, true); err != nil {
		return gear.ErrUnauthorized.From(err)
	}

	expire := time.Duration(body.Expire) * time.Second
	shareResult, err := a.models.Share.Create(EntryID, key, body.Pass, expire, &schema.Share{
		EntryID: EntryID,
		TeamID:  entry.TeamID,
		Name:    body.Name,
		UserID:  body.UserID,
	})
	if err != nil {
		return gear.ErrInternalServerError.From(err)
	}
	return ctx.JSON(200, shareResult)
}

// Delete ...
func (a *Share) Delete(ctx *gear.Context) (err error) {
	ShareID, err := util.ParseOID(ctx.Param("shareID"))
	if err != nil {
		return gear.ErrBadRequest.From(err)
	}

	userID, _ := auth.UserIDFromCtx(ctx)
	if err := a.models.Share.Delete(ShareID, userID); err != nil {
		return gear.ErrInternalServerError.From(err)
	}
	return ctx.End(204)
}
