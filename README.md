# About
The purpose of this project is to provide a simple way to set up a HTTP file server that allow you to encrypt on the fly then upload a file and to serve it ... only once! After the file is downloaded, the server deletes it for you!

## Client side
### Simple upload
* Go to **localhost:8080/upload** and you will see a simple HTML form to upload a file
* The server sends you a JSON message like this one:
``` json
{  
   "code":201,
   "path":"35709f41e0fd2f4d1ec0a3651f5944b0/file",
   "sha256sum":"15dae5979058bfbf4f9166029b6e340ea3ca374fef578a11dc9e6e923860d7ae"
}
```   
where `code` is the HTTP return code, `path` is the relative temporary path created you have to use later to download the file and `sha256sum` is the SHA256 digest of the uploaded file.
 
### Secure upload
* Go to **localhost:8080/sec-upload** to be able to encrypt your file **before** uploading it:
  * you can provide your password or let the javascript generates it for you (see `generatePassword` function in `www/static/js/upload-secu.js`)
  * Select a crypto suite to protect your file:s
    * **pbkdf2-sha256-aes-256-gcm** (recommended because it adds *integrity*)
    * **pbkdf2-sha256-aes-256-cbc**
  * if your browser has successfully encrypted your file, it uploads it
* The server sends you a JSON message like this one:
``` json
{  
   "code":201,
   "path":"1cbf9b7352e6563a1c1d4b9c785719fe/file",
   "sha256sum":"c1493207a1ae1f77e75f08810a6e877ea0d9d840bc92a8ffbf8131b6f4a0d3bd"
}
```   
where `code` is the HTTP return code, `path` is the relative temporary path created you have to use later to download the file and `sha256sum` is the SHA256 digest of the uploaded file.

### One-time download
To download a file you need its relative path returned after a successful upload, see the *one of the *upload** sections above.
* Go to **localhost:8080/dl/1cbf9b7352e6563a1c1d4b9c785719fe/file** to download it.

Be careful, you have only one try to get the file because after a GET request, the server deletes it.


## Server side
If the server is launched as *root*, it can drop its privileges to a specified user, see **configuration**. If for any reason it fails, the program stops.

# Configuration
You can edit the file `config.yml` to change default parameters.

| Option | Description | Value type | Default (config.yml)
| ------ | ----------- | ---------- | -------------------- |
| host | host | string | `"localhost"` |
| port | port | string | `8080` |
| max_size_upload_bytes | Max size of uploaded file (in bytes) | int64 | `268435456` (256MB)| 
| upload_path | Path in the file system where file will be uploaded | string | `"/var/www/share/"` |
| drop_privileges / enable | Enable privileges dropping | bool | `true` |
| drop_privileges / user_name | The name of the user who will run the program | string | `"www-data"` |
| regular_upload / enable | Enable file uploading without encryption performed by the Browser | bool | `true` |
| regular_upload / url | URL endpoint for regular upload | string | `"/upload"` |
| secure_upload / enable | Enable *secured* upload | bool | `true` |
| secure_upload / url | URL endpoint for *secured* upload | string | `"/sec-upload"` |
| download / url | URL endpoint to download files | string | `"/dl"` |

# Crypto
## Crypto Suites
Only 2 suites are available for now
### pbkdf2-sha256-aes-256-gcm
* *pbkdf2-sha256*: the password is derived with the *pbkdf2* algorithm using *SHA256* as hash function with 32000 rounds
* *aes-256-gcm*: the file is protected **both in confidentiality and integrity** with *AES-256* using *GCM* mode.
### pbkdf2-sha256-aes-256-cbc
* *pbkdf2-sha256*: the password is derived with the *pbkdf2* algorithm using *SHA256* as hash function with 32000 rounds
* *aes-256-cbc*: the file is protected **only in confidentiality** with *AES-256* using *CBC* mode

## utils/decrypt.py
A python script using [pycryptodome](https://github.com/Legrandin/pycryptodome) is provided to decipher files encrypted in the browser.

# Run it in docker
You can build a docker image using the `Dockerfile` provided:
```bash 
docker build -t alpine-golang-server .
```

Run it:
```bash
docker run -it --rm -p 8080:8080 -v /tmp/share:/var/www/share --name one-time-files-server alpine-golang-server
```

# TODO
## Client side
### Crypto
* Enable the compatibility with OpenSSL so that you won't need utils/decrypt.py script to decrypt the downloaded files.

### Server side
* secure files deletion (best effort)

## Tests
* Provide tests! 