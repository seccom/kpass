package dao

import (
	"fmt"
	"time"

	"github.com/seccom/kpass/pkg/auth"
	"github.com/seccom/kpass/pkg/schema"
	"github.com/seccom/kpass/pkg/service"
	"github.com/seccom/kpass/pkg/util"
	"github.com/teambition/gear"
	"github.com/tidwall/buntdb"
)

// Share is database access oject for share
type Share struct {
	db *service.DB
}

// NewShare return a Share intance
func NewShare(db *service.DB) *Share {
	return &Share{db}
}

// Create ...
func (o *Share) Create(EntryID util.OID, key, pass string, expire time.Duration, share *schema.Share) (
	shareResult *schema.ShareResult, err error) {
	ShareID := util.NewOID()
	token, err := auth.EncryptText(auth.SignPass(share.UserID, pass), key)
	if err != nil {
		return nil, dbError(err)
	}
	share.Token = token
	share.Created = util.Time(time.Now())
	share.Updated = share.Created
	shareResult = share.Result(ShareID)
	err = o.db.DB.Update(func(tx *buntdb.Tx) error {
		_, _, e := tx.Set(schema.ShareKey(ShareID), share.String(), &buntdb.SetOptions{
			Expires: true,
			TTL:     expire,
		})
		return e
	})
	if err != nil {
		return nil, dbError(err)
	}
	return
}

// Find ...
func (o *Share) Find(ShareID util.OID) (share *schema.Share, err error) {
	err = o.db.DB.View(func(tx *buntdb.Tx) (e error) {
		var res string
		if res, e = tx.Get(schema.ShareKey(ShareID)); e == nil {
			share, e = schema.ShareFrom(res)
		}
		return e
	})
	if err != nil {
		return nil, dbError(err)
	}
	return
}

// Delete ...
func (o *Share) Delete(ShareID util.OID, userID string) error {
	err := o.db.DB.Update(func(tx *buntdb.Tx) error {
		shareKey := schema.ShareKey(ShareID)
		value, e := tx.Get(shareKey)
		if e != nil {
			return e
		}
		share, e := schema.ShareFrom(value)
		if e != nil {
			return &gear.Error{Code: 404, Msg: "share not found"}
		}

		value, e = tx.Get(schema.TeamKey(share.TeamID))
		if e != nil {
			return e
		}
		team, e := schema.TeamFrom(value)
		if e != nil || team.IsDeleted {
			return &gear.Error{Code: 404, Msg: "team not found"}
		}
		if !team.HasMember(userID) {
			return &gear.Error{Code: 403, Msg: "not team member"}
		}
		// if team.IsFrozen {
		// 	return &gear.Error{Code: 403, Msg: "team is frozen"}
		// }
		_, e = tx.Delete(shareKey)
		return e
	})

	return dbError(err)
}

// FindByUserID ...
func (o *Share) FindByUserID(userID string) (shares []*schema.ShareResult, err error) {
	shares = make([]*schema.ShareResult, 0)
	cond := fmt.Sprintf(`{"to":"%s"}`, userID)
	err = o.db.DB.View(func(tx *buntdb.Tx) (e error) {
		tx.AscendGreaterOrEqual("share_by_user", cond, func(key, value string) bool {
			share, e := schema.ShareFrom(value)
			if e != nil {
				e = fmt.Errorf("invalid share: %s, %s", key, value)
				return false
			}
			if share.UserID != userID {
				return false
			}
			ShareID := schema.ShareIDFromKey(key)
			shares = append(shares, share.Result(ShareID))
			return true
		})
		return nil
	})
	if err != nil {
		return nil, dbError(err)
	}
	return
}

// FindByEntryID ...
func (o *Share) FindByEntryID(EntryID util.OID) (shares []*schema.ShareResult, err error) {
	shares = make([]*schema.ShareResult, 0)
	conds := fmt.Sprintf(`{"entryID":"%s"}`, EntryID.String())
	err = o.db.DB.View(func(tx *buntdb.Tx) (e error) {
		tx.AscendGreaterOrEqual("share_by_entry", conds, func(key, value string) bool {
			share, e := schema.ShareFrom(value)
			if e != nil {
				e = fmt.Errorf("invalid share: %s, %s", key, value)
				return false
			}
			if share.EntryID.String() != EntryID.String() {
				return false
			}
			ShareID := schema.ShareIDFromKey(key)
			shares = append(shares, share.Result(ShareID))
			return true
		})
		return nil
	})
	if err != nil {
		return nil, dbError(err)
	}
	return
}

// FindByTeamID ...
func (o *Share) FindByTeamID(TeamID util.OID) (shares []*schema.ShareResult, err error) {
	shares = make([]*schema.ShareResult, 0)
	conds := fmt.Sprintf(`{"teamID":"%s"}`, TeamID.String())
	err = o.db.DB.View(func(tx *buntdb.Tx) (e error) {
		tx.AscendGreaterOrEqual("share_by_team", conds, func(key, value string) bool {
			share, e := schema.ShareFrom(value)
			if e != nil {
				e = fmt.Errorf("invalid share: %s, %s", key, value)
				return false
			}
			if share.TeamID.String() != TeamID.String() {
				return false
			}
			ShareID := schema.ShareIDFromKey(key)
			shares = append(shares, share.Result(ShareID))
			return true
		})
		return nil
	})
	if err != nil {
		return nil, dbError(err)
	}
	return
}
