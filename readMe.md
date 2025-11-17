# pinata
Pinata is a simple pinata game where you throw Lemons at the Lumi pinata until it breaks, and candy falls out.
Type !throw in lumi's twitch chat to throw a lemon.


### Demo video on YouTube: 
[![Watch the video](https://img.youtube.com/vi/pR7mEA73es4/0.jpg)](https://youtu.be/pR7mEA73es4)


## library used
    - ebiten

## how to run
1. Pick the twitch channel when loading the page by appending `?channel=<name>` to the URL (for example `http://127.0.0.1:8000/dist/index.html?channel=yourchannel`). If the parameter is omitted, the game defaults to `kanekolumi`.

2. Build the game:  
    On Linux & MacOS:
    ```bash
    ./build_web.sh
    ```
    or
    ```bash
    build_web.bat
    ```
3. Run the game:  
    You will need to have a web server to run the pinata, or host it somewhere, if you have python you can run the following command to start a simple web server:
    ```bash
    python -m http.server 8000
    ```
    if you don't have python you can use any other web server.

4. Add to OBS:
    Add a browser source to the OBS scene and set the URL to the URL of the web server with the correct port, for example: http://127.0.0.1:8000

## Demo Version
[Demo Version linked to Kaneko Lumi's twitch chat](https://ca6.dev/stream/pinatalumi/)

### How to use
Include the demo version as a source browser in OBS "https://ca6.dev/stream/pinatalumi/?=yourchannel".
Once done, the pinata will react to the !throw command in Kaneko Lumi's twitch chat.
