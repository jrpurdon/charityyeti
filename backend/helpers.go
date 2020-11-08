package main

import (
	"errors"
	"net/http"

	"github.com/dghubble/go-twitter/twitter"
	"go.uber.org/zap"
)

func extractID(w http.ResponseWriter, r *http.Request) (string, error) {
	// we expect the _id of the Mongo document to come in as a query param
	id := r.URL.Query()["id"]

	// query params are found as map[string], so a length of 0 means the id param wasn't found
	if len(id) == 0 {
		return "", errors.New("no id given on request but id query parameter is required")
	}

	log.Infow("Getting record", zap.String("id", id[0]))

	return id[0], nil
}

// getInReplyToTwitterUser is a helper function takes the screen name of a Twitter user (IE - the honorary on an invoked
// tweet) and returns the full Twitter user details for that user
func getInReplyToTwitterUser(id int64) *twitterUser {
	// in the event of a retweet with comment where a user is invoking Charity Yeti, there will be no screenname so we
	// should return early
	if id == 0 {
		log.Error("No screen name provided and cannot get honorary user details")
		return nil
	}
	fullUser, _, err := twitterClient.Users.Show(&twitter.UserShowParams{
		UserID: id,
	})

	if err != nil {
		log.Errorf("cannot get honorary user details: %v", err)
	}

	user := twitterUser{
		ID:         fullUser.ID,
		Name:       fullUser.Name,
		Email:      fullUser.Email,
		ScreenName: fullUser.ScreenName,
	}
	return &user
}
