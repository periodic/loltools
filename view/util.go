package view

import (
  "appengine"
  "fmt"
  "github.com/OwenDurni/loltools/model"
  "net/http"
  "time"
)

const TIME_FORMAT = "2006-01-02 03:04PM (MST)"

// loc is the IANA Time Zone location (ex: "America/New_York")
// If the string is malformed the time is returned in UTC.
func fmtTime(t time.Time, loc string) string {
  if location, err := time.LoadLocation(loc); err != nil {
    t = t.UTC()
  } else {
    t = t.In(location)
  }
  return t.Format(TIME_FORMAT)
}

func HttpReplyOkEmpty(w http.ResponseWriter) {
  w.WriteHeader(http.StatusNoContent)
}

func HttpReplyResourceCreated(w http.ResponseWriter, loc string) {
  w.Header().Add("Location", loc)
  w.WriteHeader(http.StatusCreated)
}

func ApiHandleError(c appengine.Context, w http.ResponseWriter, err error) bool {
  useTemplate := false
  if err == nil {
    return false
  }
  if _, ok := err.(model.ErrNotAuthorized); ok {
    HttpReplyError(c, w, http.StatusForbidden, useTemplate, err)
    return true
  }
  HttpReplyError(c, w, http.StatusInternalServerError, useTemplate, err)
  return true
}

func HandleError(c appengine.Context, w http.ResponseWriter, err error) bool {
  useTemplate := true
  if err == nil {
    return false
  }
  if _, ok := err.(model.ErrNotAuthorized); ok {
    HttpReplyError(c, w, http.StatusForbidden, useTemplate, err)
    return true
  }
  HttpReplyError(c, w, http.StatusInternalServerError, useTemplate, err)
  return true
}

// See http://golang.org/pkg/net/http/#pkg-constants for status codes.
func HttpReplyError(
  c appengine.Context,
  w http.ResponseWriter,
  httpStatusCode int,
  useTemplate bool,
  err error) {

  errorString := ""
  if err != nil {
    errorString = fmt.Sprintf("%d: %s", httpStatusCode, err.Error())
  }

  // Log if this was a server-side error
  if 500 <= httpStatusCode && httpStatusCode < 600 {
    c.Errorf("%d: %s", httpStatusCode, errorString)
  }

  if !useTemplate {
    http.Error(w, err.Error(), httpStatusCode)
  } else {
    // Don't send an error code as some browsers won't render html for non-2XX responses.

    ctx := struct {
      ctxBase
      HttpStatusCode int
    }{}
    ctx.ctxBase.init(c)
    ctx.ctxBase.AddError(err)
    ctx.HttpStatusCode = httpStatusCode

    if tmplerr := RenderTemplate(w, "httperror.html", "base", ctx); tmplerr != nil {
      // Fallback to plain old response.
      http.Error(w, errorString, httpStatusCode)
    }
  }
}
