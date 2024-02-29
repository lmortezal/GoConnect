# GoConnect
**Introduction:**
***GoConnect*** is a simple GoLang project designed to facilitate the connection to a target server via SSH/SFTP protocols. It provides users with the ability to retrieve files from a specified directory on the server and write them into the local machine. This documentation provides a brief overview of the project's functionalities and usage.

### **Usage:**

```bash
./GoConnect -d <local_directory> -s <remote@server:remote_directory> -p <port>
```
**FOR EXAMPLE:**
```bash
./GoConnect -d /home/user/Downloads -s user@0.0.0.0:/home/user/Target -p 22
```

#### ***TODO:***
- [x] Add support for multiple file transfers
- [ ] Add support for file uploads
- [ ] Add recursive file transfer with directory support
- [ ] Synchonize real-time file changes
