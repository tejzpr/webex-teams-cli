[![Open Source](https://img.shields.io/badge/Open%20Source-%20-green?logo=open-source-initiative&logoColor=white&color=blue&labelColor=blue)](https://en.wikipedia.org/wiki/Open_source)
[![Golang](https://img.shields.io/badge/-Go%20Lang-blue?logo=go&logoColor=white)](https://golang.org)
[![Go Report Card](https://goreportcard.com/badge/github.com/tejzpr/webex-teams-cli)](https://goreportcard.com/report/github.com/tejzpr/webex-teams-cli)
[![Gitpod ready-to-code](https://img.shields.io/badge/Gitpod-ready--to--code-blue?logo=gitpod)](https://gitpod.io/#https://github.com/tejzpr/webex-teams-cli)

# Webex-Teams-CLI - A CLI tool to interact with Cisco Webex Teams.

## The What?
Webex Teams CLI is a versatile tool that works via Webex Teams API to interact with Webex Teams. It currently has the following features
1. Send messages
2. Broadcast to multiple rooms
3. Send Files
4. Manage a Webex team room 
5. Add / Remove users from Webex Teams room in bulk
6. A message relay server than can be hosted as a microservice allowing other services to send messages to Webex Teams.
7. Create an Onboarding Server that can onboard users via the website. 
8. Start up a Interactive Terminal UI or a Terminal console to send / recieve files or messages from Webex Teams
9. *Experimental* Start this up as a control center on a remote server and control the remote server via Webex Teams.
10. Capability to run on many architectures including Linux x86_64, Linux Arm64, Darwin (MacOs x86_64, Arm64), Windows etc. without any external dependencies. I've tested it on Alpine Linux, Raspberry PI, MacOs, Windows.  (Thanks to Golang)

## The Why?
This tool was born out of a necessity to send notifications to Webex Teams from CI/CD pipelines. Over time it has grown in capability and function therefore putting this out to the public to use / expand.

## Usage:
Set Env variable **WEBEX_ACCESS_TOKEN**, which can get retrieved from [https://developer.webex.com/docs/api/getting-started](https://developer.webex.com/docs/api/getting-started)

```sh
export WEBEX_ACCESS_TOKEN="<access_token>"
```

Send a text message (message supports markdown formatting)
```sh
webex-teams-cli room msg -t "message text" 
```

Send a text message (message supports markdown formatting) along with a file. File can be a remote http URL or a locally accessible file.
```sh
webex-teams-cli room msg -t "message text" -f <file>
```
### Using Docker:
Send a text message using the docker image
```docker
docker run -it ghcr.io/tejzpr/webex-teams-cli:main webex-teams-cli --accessToken <access-token> room --pe user@email.com msg -t "a test message"
```
Send a file _test.txt_ using docker
```docker
docker run -it -v /testdir:/testdir  ghcr.io/tejzpr/webex-teams-cli:main webex-teams-cli --accessToken <access-token> room --pe user@email.com msg -f /testdir/test.txt
```
Find the room details for a title
```docker
docker run -it ghcr.io/tejzpr/webex-teams-cli:main webex-teams-cli utils findroom -t "Room Name"
```
## Find the room details for a title
-----------------------------------------
```sh
webex-teams-cli --accessToken <access-token> utils findroom -t "Room Name"
```
## Send a message based on room ID
-----------------------------------------
Set Env variable **WEBEX_ROOM_ID**, is the Space ID that you can get by visiting https://teams.webex.com/ and clicking on a room.
```sh
export WEBEX_ROOM_ID="<roomid>"
```

Then you can send a message to the room by running the command

```sh
webex-teams-cli room msg -t "message text" 
```

**OR**

Use roomID flag to set the room ID
```sh
webex-teams-cli room -roomID <ROOMID> msg -t "message text" -f <file>
```

## Send a message to a Person based on Email address
-----------------------------------------
Set Env variable **WEBEX_PERSON_EMAIL**
```sh
export WEBEX_PERSON_EMAIL="<EMAILID>"
```

Then you can send a message to the room by running the command

```sh
webex-teams-cli room msg -t "message text" 
```

**OR**

Use toPersonEmail flag to set the person email ID
```sh
webex-teams-cli room -toPersonEmail <person@email.com> msg -t "message text" -f <file>
```

Distribution archive Includes executables for Linux amd_x64, Linux ARM5, Windows & Darwin (MacOS)

## Add Members to Room(s)
Allows to add multiple members to room(s). The member list can be passed via a .csv file with the header & data
email,moderator where email is a string and moderator acceps true/false
eg.
```people.csv
email,moderator
a@email.com,true
b@email.com,false
```
Members will be added to rooms for which you have specified permissions of either 'a' (all), 'o' (owner), 'm' (moderator) or 'om' (owner and moderator). Default is owner and moderator use the --access flag to change this.

then run the command 
```sh
webex-teams-cli room --roomID <roomID> addmembers --csv ./people.csv --access om
```
if you would like to add members to a list of rooms specified by a csv file then run
```sh
webex-teams-cli room addmembers --csv ./people.csv --roomsidscsv ./rooms.csv
```
where roomsidscsv has the following format:
```csv
roomids
<roomid-1>
<roomid-2>
```
if you would like to add members to all rooms that you have moderator access to, skip the roomID parameter
```sh
webex-teams-cli room addmembers --csv ./people.csv 
```
## Remove Members from Room(s)
Allows to remove multiple members from room(s). The member list can be passed via a .csv file with the header & data
"email" where email is a string.
eg.
```people.csv
email
a@email.com
b@email.com
```
Members will be removed from rooms for which you have specified permissions of either 'a' (all),  'o' (owner), 'm' (moderator) or 'om' (owner and moderator). Default is owner and moderator use the --access flag to change this.

then run the command 
```sh
webex-teams-cli room --roomID <roomID> removemembers --csv ./people.csv --access om
```
if you would like to remove members from a list of rooms specified by a csv file then run
```sh
webex-teams-cli room removemembers --csv ./people.csv --roomsidscsv ./rooms.csv
```
where roomsidscsv has the following format:
```csv
roomids
<roomid-1>
<roomid-2>
```
if you would like to remove members from all rooms that you have moderator access to, skip the roomID parameter
```sh
webex-teams-cli room removemembers --csv ./people.csv 
```
## Broadcast a Message or a File to a set of rooms (File broadcast will be slow, do not use for large files)
Members will be removed from rooms for which you have specified permissions of either 'a' (all),  'o' (owner), 'm' (moderator) or 'om' (owner and moderator). Default is owner and moderator use the --access flag to change this.

then run the following command to broadcast a text
```sh
webex-teams-cli room broadcast --roomsidscsv ./rooms.csv --t "message text"
```
to broadcast a file
```sh
webex-teams-cli room broadcast --roomsidscsv ./rooms.csv --f <file-path>
```
To broadcast to all rooms that you are a member of use the --access a flag
```sh
webex-teams-cli room broadcast --t "message text" --access a
```
## Shell command execution mode

**Use this mode at your own risk. This mode allows selected users system level access to remote machines through Webex**

Allows users to start up the CLI in a remote machine and then execute commands on the remote machine via Webex Teams.

**WEBEX_ACCESS_TOKEN has to be a BOT (not a USER) token to use the shell command execution mode.**

Add a bot to a Webex room and provide its access token to WEBEX_ACCESS_TOKEN, then run
```sh
webex-teams-cli shell --pes email-1@email.com,email-2@email.com --rid <roomID>  
```
Now you can interact with the BOT from the Webex Room. To view help send the *help* command from the Webex Room.
To execute a command on the remote server on which the CLI is running send the command  *cmd <command and params>* from the webex room.

## Use an interactive prompt

Allows users to start up the CLI in an interactive mode on a remote machine and communicate to Webex from the remote machine.

**WEBEX_ACCESS_TOKEN has to be a USER (not a BOT) token to use the interactive prompt.**
```sh
webex-teams-cli room --i true msg
```
Or to directly open a room
```sh
webex-teams-cli room --pe email@email.com --i true msg
```
To send a file via interactive prompt use the **sendfile** command
```
<- (email@email.com): sendfile <file-path or URI>
```
While in interactive mode the following Keyboard shortcuts are available
* Use Crtl+H to open help pane
* Use Crtl+R to change rooms (Lists most recent 50 rooms)
* TAB to switch between panes

## Use run.sh to start in interactive mode
Run the shell script run.sh and follow the prompts


## Use an console prompt (incase system doesn't support interactive mode)
Allows users to start up the CLI in an interactive (console) mode on a remote machine and communicate to Webex from the remote machine.

**WEBEX_ACCESS_TOKEN has to be a USER (not a BOT) token to use the console prompt.**
```sh
webex-teams-cli room --pe email@email.com --c true msg
```
To send a file via console prompt use the **sendfile** command
```
<- (email@email.com): sendfile <file-path or URI>
```
## Create Webex Room Onboarding Server
Create a server which can onboard users to a Webex Teams room. The user to be added is retrieved via the header **auth_user**
The email address of the user is constructed using a combination of auth_user & the emaildomain flag. The server's default port is **8000** and it can be changed by using the -p flag
```sh
webex-teams-cli adduserserver -emaildomain email.com
```
Once the server is UP a user can add themselves to a room by accessing
```
GET http://<url>/<webexroomid>
```
## Create a message relay server 
Create a message relay server send messages to a Webex Teams room 
```sh
webex-teams-cli messagerelayserver -messagerelaykey <random256lengthkey>
```
Once the server is UP a user can send messages to a room by calling the POST endpoint and sending a POST request containing the message body, and a Header **X-Message-Key** which should match the **messagerelaykey** used while starting the server.
```
POST http://<url>/<webexroomid>
```
