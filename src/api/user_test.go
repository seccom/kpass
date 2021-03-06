package api_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/DavidCai1993/request"
	"github.com/seccom/kpass/src"
	"github.com/seccom/kpass/src/auth"
	"github.com/seccom/kpass/src/schema"
	"github.com/seccom/kpass/src/util"
	"github.com/stretchr/testify/assert"
	"github.com/teambition/gear"
)

func TestUserAPI(t *testing.T) {
	app, _ := src.New("", "")
	srv := app.Start()
	defer srv.Close()

	host := "http://" + srv.Addr().String()
	id := "admin"
	pass := util.SHA256Sum(util.SHA256Sum("password"))

	t.Run("Join", func(t *testing.T) {
		assert := assert.New(t)
		user := &schema.UserResult{}

		_, err := request.Post(host+"/api/join").
			Set(gear.HeaderContentType, gear.MIMEApplicationJSON).
			Send(map[string]interface{}{"id": id, "pass": pass}).
			JSON(user)
		assert.Nil(err)

		assert.Equal(id, user.ID)
		assert.NotNil(user.Created)
		assert.Equal(user.Created, user.Updated)
	})

	t.Run("Login", func(t *testing.T) {
		assert := assert.New(t)
		res := &struct {
			ExpiresIn   int    `json:"expires_in"`
			TokenType   string `json:"token_type"`
			AccessToken string `json:"access_token"`
		}{}

		_, err := request.Post(host+"/api/login").
			Set(gear.HeaderContentType, gear.MIMEApplicationJSON).
			Send(map[string]interface{}{"username": id, "password": pass, "grant_type": "password"}).
			JSON(res)
		assert.Nil(err)

		assert.Equal(1200, res.ExpiresIn)
		assert.Equal("Bearer", res.TokenType)

		claims, _ := auth.JWT().Decode(res.AccessToken)
		assert.Equal("admin", claims.Get("id").(string))

		teams := &[]*schema.TeamResult{}

		_, err = request.Get(host+"/api/teams").
			Set(gear.HeaderAuthorization, "Bearer "+res.AccessToken).
			Set(gear.HeaderContentType, gear.MIMEApplicationJSON).
			JSON(teams)
		assert.Nil(err)
		assert.Equal(1, len(*teams))
		team := (*teams)[0]
		assert.Equal("admin", team.UserID)
		assert.Equal("private", team.Visibility)
		assert.Equal("admin", team.Members[0].ID)
	})
}

var count int = int(time.Now().Unix())

type UserInfo struct {
	ID, Pass, AccessToken, TeamID string
}

func NewUser(host string) *UserInfo {
	count++
	info := &UserInfo{}
	info.ID = "user" + strconv.Itoa(count)
	info.Pass = util.SHA256Sum(util.SHA256Sum(util.RandPass(8, 2, 2)))
	_, err := request.Post(host+"/api/join").
		Set(gear.HeaderContentType, gear.MIMEApplicationJSON).
		Send(map[string]interface{}{"id": info.ID, "pass": info.Pass}).
		End()

	if err != nil {
		panic(err)
	}

	res, err := request.Post(host+"/api/login").
		Set(gear.HeaderContentType, gear.MIMEApplicationJSON).
		Send(map[string]interface{}{"username": info.ID, "password": info.Pass, "grant_type": "password"}).
		JSON()
	if err != nil {
		panic(err)
	}

	info.AccessToken = "Bearer " + (*res.(*map[string]interface{}))["access_token"].(string)

	teams := &[]*schema.TeamResult{}
	_, err = request.Get(host+"/api/teams").
		Set(gear.HeaderAuthorization, info.AccessToken).
		Set(gear.HeaderContentType, gear.MIMEApplicationJSON).
		JSON(teams)
	if err != nil {
		panic(err)
	}

	info.TeamID = (*teams)[0].ID.String()
	return info
}
