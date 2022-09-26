# mimic
This project started off as a joke four years ago when I jokingly threatened to replace a friend of mine with a robot (I very much did). The result was a very amateur Python script that, despite being a mess, did the job quite well (although a Discord token got leaked when I pushed it to Github resulting in a massive disaster - Ouch).

Four years later, with more tools in my pocket, I decided to remake the bot to replace myself as I was preparing to deploy. I leveraged Discordgo to handle much of the API interactions, but had to write a lot of custom code and dissect the Discord API to implement all-encompassing message scraping (everyone was fine with this).

To handle the message generation, I used Markovify via Python since there were no adequate or fleshed out Go modules to do so. On startup, the Python process queries the database and builds a text model based off the user we want to simulate. Go handles the scraping, Discord API calls, and sends a small signal to the Python companion process asking for a generated message. Upon recieving the signal, the Python process uses the cached text model to generate a new message and sends it back to the Go backend using gRPC.

The app additionally is capable of converting the messages to an audio file via the Google Cloud Platform Text-To-Speech API, which it then plays in Discord using FFmpeg.

## Set-Up
1. Use `docker-compose-sample.yml` to "compose up" the mimic-app, the mimic-server Python companion, and MySQL main and replica databases.
    - ENV variables: This program operates off env variables that you must set. They are as follows...
        - mimic-sql & test-sql
            - you can leave these as default.
        - mimic
            - `TOKEN` - Set this to a Discord bot token you create yourself on the developer's portal.
            - `USER_TO_MIMIC` - Set this to the ID of a user you would like to mimic
            - `VOICE_GUILD` - Set this to the Discord server ID you want the bot to speak in.
            - `VOICE_CHANNEL` - Set this to the Discord voice channel you want the bot to speak in (I limited it so that the bot didn't get out of hand).
            - `VOICE_COMMANDER` - Set this to your own Discord user ID. Do this because GCP TTS API calls are not completely free. You only get 4 million free characters per month... Don't let the joke turn back on you when someone gets cheeky and realizes they can burn through your wallet.
            - `CHANNEL` - This env variable is deprecated. It needs to be removed in future versions. Set it to the same as your `VOICE_CHANNEL`.
            - `GRPC_SERVER` - Simply the local address of your markov chain process. If you don't change the defaults on `mimic-server` you can leave as default.
            - All DB variables can be left as default.
        - mimic-server
            - `USER_TO_MIMIC` - Set to same as mimic env variable `USER_TO_MIMIC`.

2. [OPTIONAL] From the commandline, use `sh createTopics.sh` to set up the Kafka Topics. This step is optional, as the consumers will make the topics for you.


