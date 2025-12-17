# WUZAPI-UT

WuzAPI-UT is a reimplementation of the amazing [WuzAPI](https://github.com/asternic/wuzapi) project for Ubuntu Touch.

Its a hard fork that removes most of the features, keeping the essentials to have
a Whatsapp Client on mobile devices. All credits goes to [Nicolás Gudiño](https://github.com/asternic) and contributors,
preserved in this README.

This project is designed to be a server for [Greenline](https://github.com/brennoflavio/greenline), a Whatsapp Client for Ubuntu Touch Devices.

---

<img src="static/favicon.ico" width="30"> WuzAPI is an implementation 
of the [@tulir/whatsmeow](https://github.com/tulir/whatsmeow) library as a 
simple RESTful API service with multiple device support and concurrent 
sessions.

Whatsmeow does not use Puppeteer on headless Chrome, nor an Android emulator. It communicates directly with WhatsApp’s WebSocket servers, making it significantly faster and much less demanding on memory and CPU than those solutions. The drawback is that any changes to the WhatsApp protocol could break connections, requiring a library update.

## :warning: Warning

**Using this software in violation of WhatsApp’s Terms of Service can get your number banned**:  
Be very careful—do not use this to send SPAM or anything similar. Use at your own risk. If you need to develop something for commercial purposes, contact a WhatsApp global solution provider and sign up for the WhatsApp Business API service instead.

## Available endpoints

* **Session:** Connect, disconnect, and log out from WhatsApp. Retrieve connection status and QR codes for scanning.
* **Messages:** Send text, image, audio, document, template, video, sticker, location, contact, and poll messages.
* **Users:** Check if phone numbers have WhatsApp, get user information and avatars, and retrieve the full contact list.
* **Chat:** Set presence (typing/paused, recording media), mark messages as read, download images from messages, send reactions.
* **Groups:** Create, delete and list groups, get info, get invite links, set participants, change group photos and names.

## Prerequisites

**Required:**
* Go (Go Programming Language)

**Optional:**
* Docker (for containerization)

## Updating dependencies

This project uses the whatsmeow library to communicate with WhatsApp. To update the library to the latest version, run:

```bash
go get -u go.mau.fi/whatsmeow@latest
go mod tidy
```

## Building

```
go build .
```

## Homebrew installation

To install `wuzapi` via [Homebrew](https://brew.sh) use:

```sh
brew install asternic/wuzapi/wuzapi
```

## Run

By default it will start a REST service in port 8080. These are the parameters
you can use to alter behaviour

* -address  : sets the IP address to bind the server to (default 0.0.0.0)
* -port  : sets the port number (default 8080)
* -osname : Connection OS Name in Whatsapp
* -skipmedia : Skip downloading media from messages
* -wadebug : enable whatsmeow debug, either INFO or DEBUG levels are suported

Example:

With time zone: 

Set `TZ=America/New_York ./wuzapi ...` in your shell or in your .env file or Docker Compose environment: `TZ=America/New_York`.  

## Configuration

WuzAPI uses a `.env` file for configuration. You can use the provided `.env.sample` as a template:

```bash
cp .env.sample .env
```

### Environment Variables

#### Optional Settings

```
TZ=America/New_York
SESSION_DEVICE_NAME=WuzAPI
WUZAPI_PORT=8080
```

### Important Notes

#### Database

WuzAPI uses SQLite as its database. No configuration is needed - the database files are automatically created in the `dbdata` directory.

#### Optional Settings
```
TZ=America/New_York
SESSION_DEVICE_NAME=WuzAPI
WUZAPI_PORT=8080 # Port for the WuzAPI server
```

### Docker Configuration

When using Docker Compose, `docker-compose.yml` automatically loads environment variables from a `.env` file when available. However, `docker-compose-swarm.yaml` uses `docker stack deploy`, which does not automatically load from `.env` files. Variables in the swarm file will only be substituted if they are exported in the shell environment where the deploy command is run. For managing secrets in Swarm, consider using Docker secrets.

The Docker configuration will:
1. First load variables from the `.env` file (if present and supported)
2. Use default values as fallback if variables are not defined
3. Override with any variables explicitly set in the `environment` section of the compose file

**Key notes for Docker deployment:**
- The `WUZAPI_PORT` variable controls the external port mapping in `docker-compose.yml`
- In swarm mode, `WUZAPI_PORT` configures the Traefik load balancer port
- SQLite database files are persisted in a Docker volume (`wuzapi_data`)

**Note:** The `.env` file is already included in `.gitignore` to avoid committing sensitive information to your repository.

## API reference 

API calls should be made with content type json, and parameters sent into the
request body.

Check the [API Reference](API.md)

## Contributors

<table>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/asternic>
            <img src=https://avatars.githubusercontent.com/u/25182694?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Nicolas/>
            <br />
            <sub style="font-size:14px"><b>Nicolas</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/guilhermejansen>
            <img src=https://avatars.githubusercontent.com/u/52773109?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Guilherme Jansen/>
            <br />
            <sub style="font-size:14px"><b>Guilherme Jansen</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/LuizFelipeNeves>
            <img src=https://avatars.githubusercontent.com/u/14094719?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Luiz Felipe Neves/>
            <br />
            <sub style="font-size:14px"><b>Luiz Felipe Neves</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/cleitonme>
            <img src=https://avatars.githubusercontent.com/u/12551230?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=cleitonme/>
            <br />
            <sub style="font-size:14px"><b>cleitonme</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/WellingtonFonseca>
            <img src=https://avatars.githubusercontent.com/u/25608175?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Wellington Fonseca/>
            <br />
            <sub style="font-size:14px"><b>Wellington Fonseca</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/xenodium>
            <img src=https://avatars.githubusercontent.com/u/8107219?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=xenodium/>
            <br />
            <sub style="font-size:14px"><b>xenodium</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/ramon-victor>
            <img src=https://avatars.githubusercontent.com/u/13617054?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=ramon-victor/>
            <br />
            <sub style="font-size:14px"><b>ramon-victor</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/netrixken>
            <img src=https://avatars.githubusercontent.com/u/9066682?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Netrix Ken/>
            <br />
            <sub style="font-size:14px"><b>Netrix Ken</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/luizrgf2>
            <img src=https://avatars.githubusercontent.com/u/71092163?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Luiz Ricardo Gonçalves Felipe/>
            <br />
            <sub style="font-size:14px"><b>Luiz Ricardo Gonçalves Felipe</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/andreydruz>
            <img src=https://avatars.githubusercontent.com/u/976438?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=andreydruz/>
            <br />
            <sub style="font-size:14px"><b>andreydruz</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/vitorsilvalima>
            <img src=https://avatars.githubusercontent.com/u/9752658?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Vitor Silva Lima/>
            <br />
            <sub style="font-size:14px"><b>Vitor Silva Lima</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/RuanAyram>
            <img src=https://avatars.githubusercontent.com/u/16547662?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Ruan Kaylo/>
            <br />
            <sub style="font-size:14px"><b>Ruan Kaylo</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/pedroafonso18>
            <img src=https://avatars.githubusercontent.com/u/157052926?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Pedro Afonso/>
            <br />
            <sub style="font-size:14px"><b>Pedro Afonso</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/igortrinidad>
            <img src=https://avatars.githubusercontent.com/u/13478652?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Igor Trindade/>
            <br />
            <sub style="font-size:14px"><b>Igor Trindade</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/chrsmendes>
            <img src=https://avatars.githubusercontent.com/u/77082167?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Christopher Mendes/>
            <br />
            <sub style="font-size:14px"><b>Christopher Mendes</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/luiis716>
            <img src=https://avatars.githubusercontent.com/u/97978347?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=luiis716/>
            <br />
            <sub style="font-size:14px"><b>luiis716</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/joaosouz4dev>
            <img src=https://avatars.githubusercontent.com/u/47183663?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=João Victor Souza/>
            <br />
            <sub style="font-size:14px"><b>João Victor Souza</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/gusnips>
            <img src=https://avatars.githubusercontent.com/u/981265?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Gustavo Salomé />
            <br />
            <sub style="font-size:14px"><b>Gustavo Salomé </b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/AntonKun>
            <img src=https://avatars.githubusercontent.com/u/59668952?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Anton Kozyk/>
            <br />
            <sub style="font-size:14px"><b>Anton Kozyk</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/anilgulecha>
            <img src=https://avatars.githubusercontent.com/u/1016984?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Anil Gulecha/>
            <br />
            <sub style="font-size:14px"><b>Anil Gulecha</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/AlanMartines>
            <img src=https://avatars.githubusercontent.com/u/10979090?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Alan Martines/>
            <br />
            <sub style="font-size:14px"><b>Alan Martines</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/DwiRizqiH>
            <img src=https://avatars.githubusercontent.com/u/69355492?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Ahmad Dwi Rizqi Hidayatulloh/>
            <br />
            <sub style="font-size:14px"><b>Ahmad Dwi Rizqi Hidayatulloh</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/elohmeier>
            <img src=https://avatars.githubusercontent.com/u/2536303?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=elohmeier/>
            <br />
            <sub style="font-size:14px"><b>elohmeier</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/fadlee>
            <img src=https://avatars.githubusercontent.com/u/334797?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Fadlul Alim/>
            <br />
            <sub style="font-size:14px"><b>Fadlul Alim</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/joaokopernico>
            <img src=https://avatars.githubusercontent.com/u/111400483?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=joaokopernico/>
            <br />
            <sub style="font-size:14px"><b>joaokopernico</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/JobasFernandes>
            <img src=https://avatars.githubusercontent.com/u/26033148?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Joseph Fernandes/>
            <br />
            <sub style="font-size:14px"><b>Joseph Fernandes</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/renancesarti-cyber>
            <img src=https://avatars.githubusercontent.com/u/235291917?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=renancesarti-cyber/>
            <br />
            <sub style="font-size:14px"><b>renancesarti-cyber</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/ruben18salazar3>
            <img src=https://avatars.githubusercontent.com/u/86245508?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Rubén Salazar/>
            <br />
            <sub style="font-size:14px"><b>Rubén Salazar</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/ryanachdiadsyah>
            <img src=https://avatars.githubusercontent.com/u/165612793?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Ryan Achdiadsyah/>
            <br />
            <sub style="font-size:14px"><b>Ryan Achdiadsyah</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/ViFigueiredo>
            <img src=https://avatars.githubusercontent.com/u/67883343?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=ViFigueiredo/>
            <br />
            <sub style="font-size:14px"><b>ViFigueiredo</b></sub>
        </a>
    </td>
</tr>
<tr>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/cadao7>
            <img src=https://avatars.githubusercontent.com/u/306330?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=Ricardo Maminhak/>
            <br />
            <sub style="font-size:14px"><b>Ricardo Maminhak</b></sub>
        </a>
    </td>
    <td align="center" style="word-wrap: break-word; width: 150.0; height: 150.0">
        <a href=https://github.com/zennnez>
            <img src=https://avatars.githubusercontent.com/u/3524740?v=4 width="100;"  style="border-radius:50%;align-items:center;justify-content:center;overflow:hidden;padding-top:10px" alt=zen/>
            <br />
            <sub style="font-size:14px"><b>zen</b></sub>
        </a>
    </td>
</tr>
</table>

## Clients

- [wuzapi TypeScript / Node Client](https://github.com/gusnips/wuzapi-node)

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=asternic/wuzapi&type=Date)](https://www.star-history.com/#asternic/wuzapi&Date)

## License

Copyright &copy; 2025 Nicolás Gudiño and contributors

[MIT](https://choosealicense.com/licenses/mit/)

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Icon Attribution

[Communication icons created by Vectors Market -
Flaticon](https://www.flaticon.com/free-icons/communication)

## Legal

This code is in no way affiliated with, authorized, maintained, sponsored or
endorsed by WhatsApp or any of its affiliates or subsidiaries. This is an
independent and unofficial software. Use at your own risk.

