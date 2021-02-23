# Webex-Teams-CLI - A CLI tool to send messages to / interact with Cisco Webex Teams.

![WebEx Screenshot](https://raw.githubusercontent.com/tejzpr/webex-teams-cli/main/screenshots/webex-1.png)

1. Send messages to webex teams on remote system events
2. Send message to webex periodically via cronjob
3. Use a custom bot to update room on a Jenkins event.
4. Update build stages to webex teams
5. Update intermediate binary builds paths to webex teams etc..
6. Interact with Webex via a Terminal based UI

## Usage:
Set Env variable **WEBEX_ACCESS_TOKEN**, which can get retrieved from [https://developer.webex.com/docs/api/getting-started](https://developer.webex.com/docs/api/getting-started)

```
export WEBEX_ACCESS_TOKEN="<access_token>"
```

Send a text message (message supports markdown formatting)
```
webex-teams-cli room msg -t "message text" 
```

Send a text message (message supports markdown formatting) along with a file. File can be a remote http URL or a locally accessible file.
```
webex-teams-cli room msg -t "message text" -f <file>
```

## Send a message based on room ID
-----------------------------------------
Set Env variable **WEBEX_ROOM_ID**, is the Space ID that you can get by visiting https://teams.webex.com/ and clicking on a room.
```
export WEBEX_ROOM_ID="<roomid>"
```

Then you can send a message to the room by running the command

```
webex-teams-cli room msg -t "message text" 
```

**OR**

Use roomID flag to set the room ID
```
webex-teams-cli room -roomID <ROOMID> msg -t "message text" -f <file>
```

## Send a message to a Person based on Email address
-----------------------------------------
Set Env variable **WEBEX_PERSON_EMAIL**
```
export WEBEX_PERSON_EMAIL="<EMAILID>"
```

Then you can send a message to the room by running the command

```
webex-teams-cli room msg -t "message text" 
```

**OR**

Use toPersonEmail flag to set the person email ID
```
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
```
webex-teams-cli room --roomID <roomID> addmembers --csv ./people.csv --access om
```
if you would like to add members to a list of rooms specified by a csv file then run
```
webex-teams-cli room addmembers --csv ./people.csv --roomsidscsv ./rooms.csv
```
where roomsidscsv has the following format:
```
roomids
<roomid-1>
<roomid-2>
```
if you would like to add members to all rooms that you have moderator access to, skip the roomID parameter
```
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
```
webex-teams-cli room --roomID <roomID> removemembers --csv ./people.csv --access om
```
if you would like to remove members from a list of rooms specified by a csv file then run
```
webex-teams-cli room removemembers --csv ./people.csv --roomsidscsv ./rooms.csv
```
where roomsidscsv has the following format:
```
roomids
<roomid-1>
<roomid-2>
```
if you would like to remove members from all rooms that you have moderator access to, skip the roomID parameter
```
webex-teams-cli room removemembers --csv ./people.csv 
```
## Broadcast a Message or a File to a set of rooms (File broadcast will be slow, do not use for large files)
Members will be removed from rooms for which you have specified permissions of either 'a' (all),  'o' (owner), 'm' (moderator) or 'om' (owner and moderator). Default is owner and moderator use the --access flag to change this.

then run the following command to broadcast a text
```
webex-teams-cli room broadcast --roomsidscsv ./rooms.csv --t "message text"
```
to broadcast a file
```
webex-teams-cli room broadcast --roomsidscsv ./rooms.csv --f <file-path>
```
To broadcast to all rooms that you are a member of use the --access a flag
```
webex-teams-cli room broadcast --t "message text" --access a
```
## Shell command execution mode

**Use this mode at your own risk. This mode allows selected users system level access to remote machines through Webex**

Allows users to start up the CLI in a remote machine and then execute commands on the remote machine via Webex Teams.

**WEBEX_ACCESS_TOKEN has to be a BOT (not a USER) token to use the shell command execution mode.**

Add a bot to a Webex room and provide its access token to WEBEX_ACCESS_TOKEN, then run
```
webex-teams-cli shell --pes email-1@email.com,email-2@email.com --rid <roomID>  
```
Now you can interact with the BOT from the Webex Room. To view help send the *help* command from the Webex Room.
To execute a command on the remote server on which the CLI is running send the command  *cmd <command and params>* from the webex room.

## Use an interactive prompt

Allows users to start up the CLI in an interactive mode on a remote machine and communicate to Webex from the remote machine.

**WEBEX_ACCESS_TOKEN has to be a USER (not a BOT) token to use the interactive prompt.**
```
webex-teams-cli room --i true msg
```
Or to directly open a room
```
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
```
webex-teams-cli room --pe email@email.com --c true msg
```
To send a file via console prompt use the **sendfile** command
```
<- (email@email.com): sendfile <file-path or URI>
```
## Create Webex Room Onboarding Server
Create a server which can onboard users to a Webex Teams room. The user to be added is retrieved via the header **auth_user**
The email address of the user is constructed using a combination of auth_user & the emaildomain flag. The server's default port is **8000** and it can be changed by using the -p flag
```
webex-teams-cli adduserserver -emaildomain email.com
```
Once the server is UP a user can add themselves to a room by accessing
```
GET http://<url>/<webexroomid>
```
## Create a message relay server 
Create a message relay server send messages to a Webex Teams room 
```
webex-teams-cli messagerelayserver -messagerelaykey <random256lengthkey>
```
Once the server is UP a user can send messages to a room by calling the POST endpoint and sending a POST request containing the message body, and a Header **X-Message-Key** which should match the **messagerelaykey** used while starting the server.
```
POST http://<url>/<webexroomid>
```