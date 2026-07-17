package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"time"

	"github.com/impossibleclone/fserver/internal/auth"
	"github.com/impossibleclone/fserver/internal/storage"
)

const webUITemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Secure File Server</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #121212; color: #ffffff; margin: 0; padding: 20px; }
        .container { max-width: 800px; margin: 0 auto; }
        .header { text-align: center; margin-bottom: 30px; }
        .card { background-color: #1e1e1e; border-radius: 12px; padding: 20px; margin-bottom: 20px; box-shadow: 0 4px 6px rgba(0,0,0,0.3); }
        h1 { color: #4DA8DA; margin: 0; }
        input[type="file"] { display: block; margin-bottom: 15px; color: #bbb; }
        input[type="text"] { width: 100%; padding: 12px; box-sizing: border-box; background: #2c2c2c; border: 1px solid #444; color: white; border-radius: 6px; margin-bottom: 15px; }
        button { background-color: #4DA8DA; color: #121212; border: none; padding: 10px 20px; border-radius: 6px; font-weight: bold; cursor: pointer; transition: 0.3s; }
        button:hover { background-color: #3b8ac4; }
        .btn-danger { background-color: #ff5252; color: white; padding: 6px 12px; font-size: 14px; }
        .btn-danger:hover { background-color: #e04343; }
        .file-list { list-style: none; padding: 0; margin: 0; }
        .file-item { display: flex; justify-content: space-between; align-items: center; padding: 12px; border-bottom: 1px solid #333; }
        .file-item:last-child { border-bottom: none; }
        .file-name { font-weight: 500; word-break: break-all; margin-right: 15px; }
        .file-actions { display: flex; gap: 10px; }
        .file-actions a { background-color: #4CAF50; color: white; text-decoration: none; padding: 6px 12px; border-radius: 6px; font-size: 14px; transition: 0.3s;}
        .file-actions a:hover { background-color: #45a049; }
        #adminLink { display: none; margin-top: 10px; color: #4CAF50; text-decoration: none; font-weight: bold; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Zoho SETU - File Server</h1>
            <p>Access and manage your files securely.</p>
            <a href="/admin" id="adminLink">Go to Admin Dashboard</a>
        </div>

        <div class="card">
            <h3>Upload a File</h3>
            <input type="file" id="fileInput">
            <button onclick="uploadFile()">Upload Securely</button>
        </div>

        <div class="card">
            <input type="text" id="searchInput" placeholder="Search for files..." onkeyup="filterFiles()">
            <ul class="file-list" id="fileList"></ul>
        </div>
    </div>

    <script>
        let currentFiles = [];

        async function fetchFiles() {
            const response = await fetch('/list');
            currentFiles = await response.json();
            const list = document.getElementById('fileList');
            list.innerHTML = '';
            
            if (!currentFiles || currentFiles.length === 0) {
                list.innerHTML = '<li class="file-item" style="color: #888;">No files found on server.</li>';
                return;
            }

            currentFiles.forEach(file => {
                const li = document.createElement('li');
                li.className = 'file-item';
                li.innerHTML = '<span class="file-name">' + file + '</span>' +
                    '<div class="file-actions">' +
                        '<a href="/download/' + encodeURIComponent(file) + '" download>Download</a>' +
                        '<button class="btn-danger" onclick="deleteFile(\'' + file + '\')">Delete</button>' +
                    '</div>';
                list.appendChild(li);
            });
        }

        async function uploadFile() {
            const input = document.getElementById('fileInput');
            if (input.files.length === 0) {
                alert('Please select a file to upload.');
                return;
            }
            
            const file = input.files[0];
            let overwrite = false;
            
            // Overwrite Prompt Logic
            if (currentFiles && currentFiles.includes(file.name)) {
                if (!confirm("This file already exists on the server. Do you want to overwrite it?")) {
                    return;
                }
                overwrite = true;
            }

            const formData = new FormData();
            formData.append('file', file);

            const xhr = new XMLHttpRequest();
            xhr.open('POST', '/upload?overwrite=' + overwrite, true);
            
            const startTime = new Date().getTime();
            
            let progContainer = document.getElementById('progContainer');
            if (!progContainer) {
                progContainer = document.createElement('div');
                progContainer.id = 'progContainer';
                progContainer.innerHTML = '<progress id="progBar" value="0" max="100" style="width:100%; margin-top:10px;"></progress><div id="progText" style="margin-top:5px; font-size:14px; color:#4DA8DA;"></div>';
                input.parentNode.appendChild(progContainer);
            }
            const progBar = document.getElementById('progBar');
            const progText = document.getElementById('progText');

            xhr.upload.onprogress = function(e) {
                if (e.lengthComputable) {
                    const percent = (e.loaded / e.total) * 100;
                    progBar.value = percent;
                    
                    const duration = (new Date().getTime() - startTime) / 1000;
                    let speed = "0 MB/s";
                    if (duration > 0) {
                        const bps = e.loaded / duration;
                        speed = (bps / (1024*1024)).toFixed(2) + " MB/s";
                    }
                    progText.innerText = Math.round(percent) + '% uploaded | Speed: ' + speed;
                }
            };

            xhr.onload = function() {
                if (xhr.status === 200) {
                    input.value = '';
                    progText.innerText = 'Upload complete!';
                    progBar.value = 100;
                    fetchFiles();
                } else {
                    alert('Upload failed.');
                }
            };
            xhr.send(formData);
        }

        async function deleteFile(filename) {
            if (!confirm('Are you sure you want to delete ' + filename + '?')) return;
            const response = await fetch('/delete/' + encodeURIComponent(filename), { method: 'DELETE' });
            if (response.ok) fetchFiles();
            else alert('Failed to delete file.');
        }

        function filterFiles() {
            const query = document.getElementById('searchInput').value.toLowerCase();
            const items = document.querySelectorAll('.file-item');
            items.forEach(item => {
                const name = item.querySelector('.file-name');
                if (name && name.textContent.toLowerCase().includes(query)) {
                    item.style.display = 'flex';
                } else if (name) {
                    item.style.display = 'none';
                }
            });
        }

        // Show Admin Link if user is admin
        fetch('/whoami').then(res => res.text()).then(user => {
            if (user === 'admin') {
                document.getElementById('adminLink').style.display = 'block';
            }
        });

        fetchFiles();
    </script>
</body>
</html>
`

const adminUITemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Server Admin Panel</title>
    <style>
        body { font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #121212; color: #ffffff; margin: 0; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; }
        .card { background-color: #1e1e1e; border-radius: 12px; padding: 20px; margin-bottom: 20px; }
        input { display: block; width: 100%; padding: 10px; margin-bottom: 10px; box-sizing: border-box; background: #2c2c2c; color: white; border: none; border-radius: 4px; }
        button { background-color: #4CAF50; color: white; border: none; padding: 10px 20px; border-radius: 6px; cursor: pointer; }
        button:hover { background-color: #45a049; }
        .btn-danger { background-color: #ff5252; margin-left: 10px; }
        ul { list-style: none; padding: 0; }
        li { padding: 10px; border-bottom: 1px solid #333; display: flex; justify-content: space-between; align-items: center; }
        h1 { color: #4CAF50; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Server Administration</h1>
        <p><a href="/" style="color:#4DA8DA">Back to Files</a></p>
        
        <div class="card">
            <h3>Add New User</h3>
            <input type="text" id="newUsername" placeholder="Username">
            <input type="password" id="newPassword" placeholder="Password">
            <button onclick="addUser()">Create Secure Account</button>
        </div>

        <div class="card">
            <h3>Manage Users</h3>
            <ul id="userList"></ul>
        </div>
    </div>

    <script>
        async function fetchUsers() {
            const response = await fetch('/admin/users');
            const users = await response.json();
            const list = document.getElementById('userList');
            list.innerHTML = '';
            users.forEach(user => {
                const li = document.createElement('li');
                li.innerHTML = '<span>' + user + '</span>' +
                    (user !== 'admin' ? '<button class="btn-danger" onclick="deleteUser(\'' + user + '\')">Revoke</button>' : '');
                list.appendChild(li);
            });
        }

        async function addUser() {
            const u = document.getElementById('newUsername').value;
            const p = document.getElementById('newPassword').value;
            if (!u || !p) return;

            const response = await fetch('/admin/users?username=' + encodeURIComponent(u) + '&password=' + encodeURIComponent(p), { method: 'POST' });
            if (response.ok) {
                document.getElementById('newUsername').value = '';
                document.getElementById('newPassword').value = '';
                fetchUsers();
            } else {
                alert('Failed to add user');
            }
        }

        async function deleteUser(username) {
            if (!confirm('Revoke access for ' + username + '?')) return;
            await fetch('/admin/users?username=' + encodeURIComponent(username), { method: 'DELETE' });
            fetchUsers();
        }

        fetchUsers();
    </script>
</body>
</html>
`

type Handler struct {
	vfs  storage.VFS
	auth auth.Authenticator
}

func NewHandler(vfs storage.VFS, authenticator auth.Authenticator) *Handler {
	return &Handler{
		vfs:  vfs,
		auth: authenticator,
	}
}

func (h *Handler) HandleWhoAmI(w http.ResponseWriter, r *http.Request) {
	u, _, _ := r.BasicAuth()
	w.Write([]byte(u))
}

func (h *Handler) HandleWebUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(webUITemplate))
}

func (h *Handler) HandleAdminUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(adminUITemplate))
}

func (h *Handler) HandleAdminUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		users := h.auth.GetUsers()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
		return
	}
	
	if r.Method == http.MethodPost {
		u := r.URL.Query().Get("username")
		p := r.URL.Query().Get("password")
		if err := h.auth.AddUser(u, p); err != nil {
			http.Error(w, "Failed to add user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	
	if r.Method == http.MethodDelete {
		u := r.URL.Query().Get("username")
		if u == "admin" {
			http.Error(w, "Cannot delete admin", http.StatusBadRequest)
			return
		}
		h.auth.RemoveUser(u)
		w.WriteHeader(http.StatusOK)
		return
	}
}

func (h *Handler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	overwrite := r.URL.Query().Get("overwrite") == "true"

	if err := h.vfs.SaveFile(filename, file, overwrite); err != nil {
		http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := filepath.Base(r.URL.Path)
	file, err := h.vfs.GetFile(filename)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	http.ServeContent(w, r, filename, time.Now(), file)
}

func (h *Handler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	filename := filepath.Base(r.URL.Path)
	if err := h.vfs.DeleteFile(filename); err != nil {
		http.Error(w, "Failed to delete file", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	files, err := h.vfs.ListFiles()
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}
