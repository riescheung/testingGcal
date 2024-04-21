package main

import (
    "fmt"
    "net/http"
    "os"
    "time"

    "golang.org/x/net/context"
    "golang.org/x/oauth2"
    "golang.org/x/oauth2/google"
    "google.golang.org/api/calendar/v3"
    "google.golang.org/api/drive/v3"
)

var googleOauthConfig = &oauth2.Config{
    RedirectURL:  "http://localhost:3000/GoogleCallback",
    ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
    ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
    Scopes:       []string{"https://www.googleapis.com/auth/drive", "https://www.googleapis.com/auth/calendar"},
    Endpoint:     google.Endpoint,
}

var oauthStateString = "pseudo-random"

const htmlIndex = `<html><body>
<a href="/GoogleLogin">Log in with Google</a>
</body></html>
`

func main() {
    http.HandleFunc("/", handleMain)
    http.HandleFunc("/GoogleLogin", handleGoogleLogin)
    http.HandleFunc("/GoogleCallback", handleGoogleCallback)
    fmt.Println("Server started at http://localhost:3000")
    fmt.Println(http.ListenAndServe(":3000", nil))
}

func handleMain(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, htmlIndex)
}

func handleGoogleLogin(w http.ResponseWriter, r *http.Request) {
    url := googleOauthConfig.AuthCodeURL(oauthStateString)
    http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleGoogleCallback(w http.ResponseWriter, r *http.Request) {
    if r.FormValue("state") != oauthStateString {
        fmt.Fprintf(w, "Invalid OAuth state")
        return
    }

    code := r.FormValue("code")
    token, err := googleOauthConfig.Exchange(context.Background(), code)
    if err != nil {
        fmt.Fprintf(w, "OAuth exchange failed: %v", err)
        return
    }

    client := oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(token))
    calendarService, err := calendar.New(client)
    if err != nil {
        fmt.Fprintf(w, "Calendar client creation failed: %v", err)
        return
    }

    driveService, err := drive.New(client)
    if err != nil {
        fmt.Fprintf(w, "Drive client creation failed: %v", err)
        return
    }

    calendarEvents, err := calendarService.Events.List("primary").TimeMin(time.Now().Format(time.RFC3339)).MaxResults(5).Do()
    if err != nil {
        fmt.Fprintf(w, "Error retrieving calendar events: %v", err)
        return
    }

    if len(calendarEvents.Items) > 0 {
        for _, i := range calendarEvents.Items {
            fmt.Fprintf(w, "Event: %s at %s<br>", i.Summary, i.Start.DateTime)
        }
    } else {
        fmt.Fprintf(w, "No upcoming events found.<br>")
    }

    fileList, err := driveService.Files.List().PageSize(10).Fields("nextPageToken, files(id, name)").Do()
    if err != nil {
        fmt.Fprintf(w, "Error retrieving drive files: %v", err)
        return
    }

    if len(fileList.Files) > 0 {
        for _, i := range fileList.Files {
            fmt.Fprintf(w, "File: %s (%s)<br>", i.Name, i.Id)
        }
    } else {
        fmt.Fprintf(w, "No files found.<br>")
    }
}
