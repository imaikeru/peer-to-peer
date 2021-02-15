# Peer-to-Peer

A simple peer-to-peer file exchange system, implemented in Go.

## Download/Install

To download the project, you have to clone it:
```
https://github.com/imaikeru/peer-to-peer.git
```

## Running
### After you have downloaded the project
Go to project directory:
```
cd peer-to-peer
```
### 1. Start Server
From project directory:
```
cd server
go run main.go
````

### 2. Start Client
From project directory:
```
cd client
go run main.go -file_path="/Absolute/Path/To/Existing/File/Where/Usernames/And/Addresses/Will/Be/Saved"
```

## Usage - On Client
**To announce which files are available for downloading from you:**
```
register username "file1" "file2" "file3" ... "fileN"
```

**To announce which files are *NO LONGER* available for downloading from you:**
```
unregister username "file1" "file2" "file3" ... "fileN"
```
**To list available files and the users they are available to download from:**
```
list-files
```
**Тo download a file from another user:**
```
download otheruser "/absolute/path/to/file/on/other/user" "/absolute/path/to/save/on/current/user"
```
**To disconnect from server:**
```
disconnect
```
## Example - Client
```
go run main.go -file_path="D:\myFiles\usersInfo.txt"

register gosho "D:\myFiles\lyrics.txt" "D:\myFiles\mydoc.txt"

list-files

unregister gosho "D:\myFiles\mydoc.txt"

download pesho "E:\files\parola.txt" "D:\myFiles\peshoPassword.txt"

disconnect 
```