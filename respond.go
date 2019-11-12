package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/dghubble/go-twitter/twitter"
)

// respondToTweet receives an incoming tweet from a stream,
// and will respond to it with a link to donate to either
// Foundation To Decrease World Suck or directly at Parners
// In Health
func respondToTweet(username string, honorary string, tweetID int64) error {
	if honorary != "" {
		donateLink := "https://donate.pih.org/page/contribute/maternal-health-sierra-leone"
		tweetText := fmt.Sprintf("Hi @%s! You can donate to PiH on @%s's behalf here: %s", username, honorary, donateLink)

		if sendResponses {
			log.Print("Actually sending this!")
			_, _, err := client.Statuses.Update(tweetText, nil)
			if err != nil {
				return err
			}
		}

		log.Print(tweetText)

		return nil
	}

	return errors.New("No honorary to respond to")
}

// generateResponse parses the tweet from the demux stream, pulls out
// the user who sent it, the ID of the originating tweet, and
// passes it to respondToTweet to send to Twitter
func generateResponse(incomingTweet *twitter.Tweet) {
	user := incomingTweet.User
	honorary := incomingTweet.InReplyToScreenName
	tweetID := incomingTweet.ID

	err := respondToTweet(user.ScreenName, honorary, tweetID)

	if err != nil {
		log.Printf(err.Error())
	}
}
