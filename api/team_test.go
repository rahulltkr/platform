// Copyright (c) 2015 Mattermost, Inc. All Rights Reserved.
// See License.txt for license information.

package api

import (
	"fmt"
	"github.com/mattermost/platform/model"
	"github.com/mattermost/platform/store"
	"github.com/mattermost/platform/utils"
	"strings"
	"testing"
)

func TestSignupTeam(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	_, err := Client.SignupTeam("test@nowhere.com", "name")
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateFromSignupTeam(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	props := make(map[string]string)
	props["email"] = strings.ToLower(model.NewId()) + "success+test@simulator.amazonses.com"
	props["name"] = "Test Company name"
	props["time"] = fmt.Sprintf("%v", model.GetMillis())

	data := model.MapToJson(props)
	hash := model.HashPassword(fmt.Sprintf("%v:%v", data, utils.Cfg.EmailSettings.InviteSalt))

	team := model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
	user := model.User{Email: props["email"], Nickname: "Corey Hulen", Password: "hello"}

	ts := model.TeamSignup{Team: team, User: user, Invites: []string{"success+test@simulator.amazonses.com"}, Data: data, Hash: hash}

	rts, err := Client.CreateTeamFromSignup(&ts)
	if err != nil {
		t.Fatal(err)
	}

	if rts.Data.(*model.TeamSignup).Team.DisplayName != team.DisplayName {
		t.Fatal("full name didn't match")
	}

	ruser := rts.Data.(*model.TeamSignup).User
	rteam := rts.Data.(*model.TeamSignup).Team
	Client.SetTeamId(rteam.Id)

	if result, err := Client.LoginById(ruser.Id, user.Password); err != nil {
		t.Fatal(err)
	} else {
		if result.Data.(*model.User).Email != user.Email {
			t.Fatal("email's didn't match")
		}
	}

	c1 := Client.Must(Client.GetChannels("")).Data.(*model.ChannelList)
	if len(c1.Channels) != 2 {
		t.Fatal("default channels not created")
	}

	ts.Data = "garbage"
	_, err = Client.CreateTeamFromSignup(&ts)
	if err == nil {
		t.Fatal(err)
	}
}

func TestCreateTeam(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
	rteam, err := Client.CreateTeam(&team)
	if err != nil {
		t.Fatal(err)
	}

	user := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	user = Client.Must(Client.CreateUser(user, "")).Data.(*model.User)
	LinkUserToTeam(user, rteam.Data.(*model.Team))
	store.Must(Srv.Store.User().VerifyEmail(user.Id))

	Client.LoginByEmail(team.Name, user.Email, "pwd")
	Client.SetTeamId(rteam.Data.(*model.Team).Id)

	c1 := Client.Must(Client.GetChannels("")).Data.(*model.ChannelList)
	if len(c1.Channels) != 2 {
		t.Fatal("default channels not created")
	}

	if rteam.Data.(*model.Team).DisplayName != team.DisplayName {
		t.Fatal("full name didn't match")
	}

	if _, err := Client.CreateTeam(rteam.Data.(*model.Team)); err == nil {
		t.Fatal("Cannot create an existing")
	}

	rteam.Data.(*model.Team).Id = ""
	if _, err := Client.CreateTeam(rteam.Data.(*model.Team)); err != nil {
		if err.Message != "A team with that domain already exists" {
			t.Fatal(err)
		}
	}

	if _, err := Client.DoApiPost("/teams/create", "garbage"); err == nil {
		t.Fatal("should have been an error")
	}
}

func TestAddUserToTeam(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	props := make(map[string]string)
	props["email"] = strings.ToLower(model.NewId()) + "success+test@simulator.amazonses.com"
	props["name"] = "Test Company name"
	props["time"] = fmt.Sprintf("%v", model.GetMillis())

	data := model.MapToJson(props)
	hash := model.HashPassword(fmt.Sprintf("%v:%v", data, utils.Cfg.EmailSettings.InviteSalt))

	team := model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: props["email"], Type: model.TEAM_OPEN}
	user := model.User{Email: props["email"], Nickname: "Corey Hulen", Password: "hello"}

	ts := model.TeamSignup{Team: team, User: user, Invites: []string{"success+test@simulator.amazonses.com"}, Data: data, Hash: hash}

	rts, err := Client.CreateTeamFromSignup(&ts)
	if err != nil {
		t.Fatal(err)
	}

	if rts.Data.(*model.TeamSignup).Team.DisplayName != team.DisplayName {
		t.Fatal("full name didn't match")
	}

	ruser := rts.Data.(*model.TeamSignup).User
	rteam := rts.Data.(*model.TeamSignup).Team
	Client.SetTeamId(rteam.Id)

	if result, err := Client.LoginById(ruser.Id, user.Password); err != nil {
		t.Fatal(err)
	} else {
		if result.Data.(*model.User).Email != user.Email {
			t.Fatal("email's didn't match")
		}
	}

	user2 := th.CreateUser(th.BasicClient)
	if result, err := th.BasicClient.AddUserToTeam(user2.Id); err != nil {
		t.Fatal(err)
	} else {
		rm := result.Data.(map[string]string)
		if rm["user_id"] != user2.Id {
			t.Fatal("email's didn't match")
		}
	}
}

func TestAddUserToTeamFromInvite(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	props := make(map[string]string)
	props["email"] = strings.ToLower(model.NewId()) + "success+test@simulator.amazonses.com"
	props["name"] = "Test Company name"
	props["time"] = fmt.Sprintf("%v", model.GetMillis())

	data := model.MapToJson(props)
	hash := model.HashPassword(fmt.Sprintf("%v:%v", data, utils.Cfg.EmailSettings.InviteSalt))

	team := model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: props["email"], Type: model.TEAM_OPEN}
	user := model.User{Email: props["email"], Nickname: "Corey Hulen", Password: "hello"}

	ts := model.TeamSignup{Team: team, User: user, Invites: []string{"success+test@simulator.amazonses.com"}, Data: data, Hash: hash}

	rts, err := Client.CreateTeamFromSignup(&ts)
	if err != nil {
		t.Fatal(err)
	}

	if rts.Data.(*model.TeamSignup).Team.DisplayName != team.DisplayName {
		t.Fatal("full name didn't match")
	}

	ruser := rts.Data.(*model.TeamSignup).User
	rteam := rts.Data.(*model.TeamSignup).Team
	Client.SetTeamId(rteam.Id)

	if result, err := Client.LoginById(ruser.Id, user.Password); err != nil {
		t.Fatal(err)
	} else {
		if result.Data.(*model.User).Email != user.Email {
			t.Fatal("email's didn't match")
		}
	}

	user2 := th.CreateUser(th.BasicClient)
	Client.Must(Client.Logout())
	Client.Must(Client.LoginByEmail("", user2.Email, user2.Password))

	if result, err := th.BasicClient.AddUserToTeamFromInvite("", "", rteam.InviteId); err != nil {
		t.Fatal(err)
	} else {
		rtm := result.Data.(*model.Team)
		if rtm.Id != rteam.Id {
			t.Fatal()
		}
	}
}

func TestGetAllTeams(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
	team = Client.Must(Client.CreateTeam(team)).Data.(*model.Team)

	user := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	user = Client.Must(Client.CreateUser(user, "")).Data.(*model.User)
	LinkUserToTeam(user, team)
	store.Must(Srv.Store.User().VerifyEmail(user.Id))

	Client.LoginByEmail(team.Name, user.Email, "pwd")
	Client.SetTeamId(team.Id)

	if r1, err := Client.GetAllTeams(); err != nil {
		t.Fatal(err)
	} else {
		teams := r1.Data.(map[string]*model.Team)
		if teams[team.Id].Name != team.Name {
			t.Fatal()
		}
		if teams[team.Id].Email != "" {
			t.Fatal("Non admin users shoudn't get full listings")
		}
	}

	c := &Context{}
	c.RequestId = model.NewId()
	c.IpAddress = "cmd_line"
	UpdateRoles(c, user, model.ROLE_SYSTEM_ADMIN)

	Client.LoginByEmail(team.Name, user.Email, "pwd")
	Client.SetTeamId(team.Id)

	if r1, err := Client.GetAllTeams(); err != nil {
		t.Fatal(err)
	} else {
		teams := r1.Data.(map[string]*model.Team)
		if teams[team.Id].Name != team.Name {
			t.Fatal()
		}
		if teams[team.Id].Email != team.Email {
			t.Fatal()
		}
	}
}

func TestGetAllTeamListings(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN, AllowOpenInvite: true}
	team = Client.Must(Client.CreateTeam(team)).Data.(*model.Team)

	user := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	user = Client.Must(Client.CreateUser(user, "")).Data.(*model.User)
	LinkUserToTeam(user, team)
	store.Must(Srv.Store.User().VerifyEmail(user.Id))

	Client.LoginByEmail(team.Name, user.Email, "pwd")
	Client.SetTeamId(team.Id)

	if r1, err := Client.GetAllTeamListings(); err != nil {
		t.Fatal(err)
	} else {
		teams := r1.Data.(map[string]*model.Team)
		if teams[team.Id].Name != team.Name {
			t.Fatal()
		}
		if teams[team.Id].Email != "" {
			t.Fatal("Non admin users shoudn't get full listings")
		}
	}

	c := &Context{}
	c.RequestId = model.NewId()
	c.IpAddress = "cmd_line"
	UpdateRoles(c, user, model.ROLE_SYSTEM_ADMIN)

	Client.LoginByEmail(team.Name, user.Email, "pwd")
	Client.SetTeamId(team.Id)

	if r1, err := Client.GetAllTeams(); err != nil {
		t.Fatal(err)
	} else {
		teams := r1.Data.(map[string]*model.Team)
		if teams[team.Id].Name != team.Name {
			t.Fatal()
		}
		if teams[team.Id].Email != team.Email {
			t.Fatal()
		}
	}
}

func TestTeamPermDelete(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
	team = Client.Must(Client.CreateTeam(team)).Data.(*model.Team)

	user1 := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	user1 = Client.Must(Client.CreateUser(user1, "")).Data.(*model.User)
	LinkUserToTeam(user1, team)
	store.Must(Srv.Store.User().VerifyEmail(user1.Id))

	Client.LoginByEmail(team.Name, user1.Email, "pwd")
	Client.SetTeamId(team.Id)

	channel1 := &model.Channel{DisplayName: "TestGetPosts", Name: "a" + model.NewId() + "a", Type: model.CHANNEL_OPEN, TeamId: team.Id}
	channel1 = Client.Must(Client.CreateChannel(channel1)).Data.(*model.Channel)

	post1 := &model.Post{ChannelId: channel1.Id, Message: "search for post1"}
	post1 = Client.Must(Client.CreatePost(post1)).Data.(*model.Post)

	post2 := &model.Post{ChannelId: channel1.Id, Message: "search for post2"}
	post2 = Client.Must(Client.CreatePost(post2)).Data.(*model.Post)

	post3 := &model.Post{ChannelId: channel1.Id, Message: "#hashtag search for post3"}
	post3 = Client.Must(Client.CreatePost(post3)).Data.(*model.Post)

	post4 := &model.Post{ChannelId: channel1.Id, Message: "hashtag for post4"}
	post4 = Client.Must(Client.CreatePost(post4)).Data.(*model.Post)

	c := &Context{}
	c.RequestId = model.NewId()
	c.IpAddress = "test"

	err := PermanentDeleteTeam(c, team)
	if err != nil {
		t.Fatal(err)
	}

	Client.ClearOAuthToken()
}

func TestInviteMembers(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
	team = Client.Must(Client.CreateTeam(team)).Data.(*model.Team)

	user := &model.User{Email: model.NewId() + "success+test@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	user = Client.Must(Client.CreateUser(user, "")).Data.(*model.User)
	LinkUserToTeam(user, team)
	store.Must(Srv.Store.User().VerifyEmail(user.Id))

	Client.LoginByEmail(team.Name, user.Email, "pwd")
	Client.SetTeamId(team.Id)

	invite := make(map[string]string)
	invite["email"] = "success+" + model.NewId() + "@simulator.amazonses.com"
	invite["first_name"] = "Test"
	invite["last_name"] = "Guy"
	invites := &model.Invites{Invites: []map[string]string{invite}}
	invites.Invites = append(invites.Invites, invite)

	if _, err := Client.InviteMembers(invites); err != nil {
		t.Fatal(err)
	}

	invites = &model.Invites{Invites: []map[string]string{}}
	if _, err := Client.InviteMembers(invites); err == nil {
		t.Fatal("Should have errored out on no invites to send")
	}
}

func TestUpdateTeamDisplayName(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "success+" + model.NewId() + "@simulator.amazonses.com", Type: model.TEAM_OPEN}
	team = Client.Must(Client.CreateTeam(team)).Data.(*model.Team)

	user := &model.User{Email: team.Email, Nickname: "Corey Hulen", Password: "pwd"}
	user = Client.Must(Client.CreateUser(user, "")).Data.(*model.User)
	LinkUserToTeam(user, team)
	store.Must(Srv.Store.User().VerifyEmail(user.Id))

	user2 := &model.User{Email: "success+" + model.NewId() + "@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	user2 = Client.Must(Client.CreateUser(user2, "")).Data.(*model.User)
	LinkUserToTeam(user2, team)
	store.Must(Srv.Store.User().VerifyEmail(user2.Id))

	Client.LoginByEmail(team.Name, user2.Email, "pwd")
	Client.SetTeamId(team.Id)

	vteam := &model.Team{DisplayName: team.DisplayName, Name: team.Name, Email: team.Email, Type: team.Type}
	vteam.DisplayName = "NewName"
	if _, err := Client.UpdateTeam(vteam); err == nil {
		t.Fatal("Should have errored, not admin")
	}

	Client.LoginByEmail(team.Name, user.Email, "pwd")

	vteam.DisplayName = ""
	if _, err := Client.UpdateTeam(vteam); err == nil {
		t.Fatal("Should have errored, empty name")
	}

	vteam.DisplayName = "NewName"
	if _, err := Client.UpdateTeam(vteam); err != nil {
		t.Fatal(err)
	}
}

func TestFuzzyTeamCreate(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	for i := 0; i < len(utils.FUZZY_STRINGS_NAMES) || i < len(utils.FUZZY_STRINGS_EMAILS); i++ {
		testDisplayName := "Name"
		testEmail := "test@nowhere.com"

		if i < len(utils.FUZZY_STRINGS_NAMES) {
			testDisplayName = utils.FUZZY_STRINGS_NAMES[i]
		}
		if i < len(utils.FUZZY_STRINGS_EMAILS) {
			testEmail = utils.FUZZY_STRINGS_EMAILS[i]
		}

		team := model.Team{DisplayName: testDisplayName, Name: "z-z-" + model.NewId() + "a", Email: testEmail, Type: model.TEAM_OPEN}

		_, err := Client.CreateTeam(&team)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestGetMyTeam(t *testing.T) {
	th := Setup().InitBasic()
	th.BasicClient.Logout()
	Client := th.BasicClient

	team := &model.Team{DisplayName: "Name", Name: "z-z-" + model.NewId() + "a", Email: "test@nowhere.com", Type: model.TEAM_OPEN}
	rteam, _ := Client.CreateTeam(team)
	team = rteam.Data.(*model.Team)

	user := model.User{Email: "success+" + model.NewId() + "@simulator.amazonses.com", Nickname: "Corey Hulen", Password: "pwd"}
	ruser, _ := Client.CreateUser(&user, "")
	LinkUserToTeam(ruser.Data.(*model.User), rteam.Data.(*model.Team))
	store.Must(Srv.Store.User().VerifyEmail(ruser.Data.(*model.User).Id))

	Client.LoginByEmail(team.Name, user.Email, user.Password)
	Client.SetTeamId(team.Id)

	if result, err := Client.GetMyTeam(""); err != nil {
		t.Fatal(err)
	} else {
		if result.Data.(*model.Team).DisplayName != team.DisplayName {
			t.Fatal("team names did not match")
		}
		if result.Data.(*model.Team).Name != team.Name {
			t.Fatal("team domains did not match")
		}
		if result.Data.(*model.Team).Type != team.Type {
			t.Fatal("team types did not match")
		}
	}
}

func TestGetTeamMembers(t *testing.T) {
	th := Setup().InitBasic()

	if result, err := th.BasicClient.GetTeamMembers(th.BasicTeam.Id); err != nil {
		t.Fatal(err)
	} else {
		members := result.Data.([]*model.TeamMember)
		t.Log(members)
	}
}
