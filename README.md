# Voice UI

This is a demo of using Twilio and Google Speech API to build your own custom speech recognition service.

It is an implementation of the [JustForFunc "magic-gate" tutorial](https://youtu.be/mTd3hHUy9OU).

## Notes:

This demo is designed to run on Google App Engine. Simple modifications to the code would allow one to make a normal static binary that could be deployed anywhere. Perhaps I will create a standalone branch for that purpose.

For now, we are strictly using GAE, as in Francesc's tutorial.

You will need a Google Speech API key. Place your API key in `app.yaml`

## To run locally:

```
$ cd voice-ui
$ goapp serve
```

## To deploy to GAE:

```
$ cd voice-ui
$ goapp deploy --application=<yourprojecthere> --version=<yourversionhere>
```

### Other notes:

While the JustForFunc tutorial is, of course, about magically opening the building gate, my code simply checks whether the user has supplied the right password. Otherwise, it is possible to follow along using this repo.

As is mentioned in the tutorial, an easier route for hooking up both external services during development is to use [ngrok](http://ngrok.io).

You can then just change the Twilio endpoint to your ngrok address, and later switch it over when you deploy the app.

Speech transcriptions are parsed in the main handler function, but for more complex IVR-like decision trees one would probably want to break that out into its own function 

## License

(C) 2016 by Damien Stanton, with attribution to Francesc Campoy

Free to use and modify per the MIT License.