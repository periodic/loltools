package view

import (
  "appengine"
  "appengine/datastore"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
)

type League struct {
  Name string
  Owner string
  Id string
  Uri string
}
func (l *League) Fill(m *model.League, k *datastore.Key) *League {
  l.Name = m.Name
  l.Id = model.EncodeKeyShort(k)
  l.Uri = model.LeagueUri(k)
  return l
}

type Team struct {
  Name string
  Id string
  Uri string
  
  Wins int
  Losses int
}
func (t *Team) Fill(m *model.TeamInfo) *Team {
  t.Name = m.Name
  t.Id = m.Id
  t.Uri = m.Uri
  return t
}

func LeagueIndexHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)

  // Lookup data from backend.
  _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  leagueInfos, err := model.LeaguesForUser(c, userKey)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  // Populate view context.
  ctx := struct {
    ctxBase
    MyLeagues []*League
  }{}
  ctx.ctxBase.init(c)
  
  ctx.MyLeagues = make([]*League, len(leagueInfos))
  for i, info := range leagueInfos {
    league := new(League).Fill(&info.League, info.LeagueKey)
    if owner, err := model.GetUserByKey(c, info.League.Owner); err == nil {
      league.Owner = owner.Email
    } else {
      league.Owner = err.Error()
    }
    ctx.MyLeagues[i] = league
  }
  
  // Render
  if err := RenderTemplate(w, "leagues/index.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func LeagueViewHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := args["leagueId"]

  _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  league, leagueKey, err := model.LeagueById(c, userKey, leagueId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  teamInfos, err := model.LeagueAllTeams(c, userKey, leagueKey)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  // Populate view context.
  ctx := struct {
    ctxBase
    League
    Teams []Team
  }{}
  ctx.ctxBase.init(c)
  ctx.ctxBase.Title = fmt.Sprintf("loltools - %s", league.Name)
  
  ctx.League.Fill(league, leagueKey)
  
  ctx.Teams = make([]Team, len(teamInfos))
  for i, _ := range ctx.Teams {
    ctx.Teams[i].Fill(teamInfos[i])
  }
  
  // Render
  if err := RenderTemplate(w, "leagues/view.html", "base", ctx); err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
}

func ApiLeagueCreateHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  _, leagueKey, err := model.CreateLeague(c, r.FormValue("name"))
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  HttpReplyResourceCreated(w, model.LeagueUri(leagueKey))
}

func ApiLeagueAddTeamHandler(w http.ResponseWriter, r *http.Request, args map[string]string) {
  c := appengine.NewContext(r)
  leagueId := r.FormValue("league")
  teamName := r.FormValue("team")
  
  _, userKey, err := model.GetUser(c)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  _, leagueKey, err := model.LeagueById(c, userKey, leagueId)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  _, teamKey, err := model.LeagueAddTeam(c, userKey, leagueId, teamName)
  if err != nil {
    HttpReplyError(w, r, http.StatusInternalServerError, err)
    return
  }
  
  HttpReplyResourceCreated(w, model.LeagueTeamUri(leagueKey, teamKey))
}