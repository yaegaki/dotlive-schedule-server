package tweet

import (
	"os"

	"github.com/ChimeraCoder/anaconda"
)

func init() {
	anaconda.SetConsumerKey(os.Getenv("TWITTER_CONSUMER_KEY"))
	anaconda.SetConsumerSecret(os.Getenv("TWITTER_CONSUMER_SECRET"))
}
