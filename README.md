# MetricRain
This bot converts Tweets containing rain measurements from inches to cm.

This bot responds to Tweets from [@SummerhillRain](https://twitter.com/SummerhillRain) and uses the handle
[@MetricMartin](https://twitter.com/MetricMartin/with_replies). 

This code should be put into a Google Cloud Function. Set up a cron job to call the
function's webhook every 15 minutes. On each run it will respond to any Tweets that it 
has not yet responded to, that contain a rain measurement.