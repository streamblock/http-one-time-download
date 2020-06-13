const inputFile = document.getElementById("uploadFile");
inputFile.addEventListener("change", updateNameAndSize, false);
const openssl = document.getElementById("switchOpenssl");
const cryptoSuite = document.getElementById("selectCryptoAlgo");
const password = document.getElementById("passwordInput");
const uploadFileStatus = document.getElementById("uploadFileStatus");

const generatePasswordButton = document.getElementById("generatePasswordButton");
generatePasswordButton.addEventListener("click", function(){generatePassword(getCryptoSuite())}, false);

const uploadButton = document.getElementById("uploadButton");
uploadButton.addEventListener("click", encryptThenUpload);

const resetButton = document.getElementById("resetButton");
resetButton.addEventListener("click", resetInputs);

/*********** Crypto parameters *********/
// Crypto Suites
const CRYPTO_SUITE_PBKDF2_SHA256_AES_256_GCM = {
    cryptoSuiteName: "pbkdf2-sha256-aes-256-gcm",
    pwdLength: 16,
    kdfAlgoName: "PBKDF2",
    kdfIterations: 32000,
    kdfHashFunctionName: "SHA-256",
    encryptAlgoName: "AES-GCM",
    encryptKeySizeBits: 256,
    encryptAlgoBlockBitsSize:128
};

const CRYPTO_SUITE_PBKDF2_SHA256_AES_256_CBC = {
    cryptoSuiteName:"pbkdf2-sha256-aes-256-cbc",
    pwdLength:16,
    kdfAlgoName: "PBKDF2",
    kdfIterations: 32000,
    kdfHashFunctionName: "SHA-256",
    encryptAlgoName: "AES-CBC",
    encryptKeyLengthBits: 256,
    encryptAlgoBlockBitsSize:128
};

const OPENSSL_HEADER_FILE = "Salted__";

/*********** Crypto functions *********/
function generatePassword(cryptoSuite) {
    showResetButton();
    const usedChars = '0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!"#$%&\'()*+,-./:;<=>?@[\\]^_`{|}~';
    let keyArray = new Uint8Array(cryptoSuite.pwdLength);
    window.crypto.getRandomValues(keyArray);
    keyArray = keyArray.map(x => usedChars.charCodeAt(x % usedChars.length));
    const randomizedKey = String.fromCharCode.apply(null, keyArray);
    console.log("[+] password generated: ", randomizedKey);
    setPasswordInput(randomizedKey);
}

function setPasswordInput(passwordInput) {
    password.value = passwordInput;
}

function getPasswordFromInput() {
    if (!password.value) {
        errorMessage("Please enter your password or generate one")
        return;
    } else {
        let encoder = new TextEncoder('utf-8');
        return encoder.encode(password.value);
    }
}

function getCryptoSuite() {
    var cryptoSuiteName = cryptoSuite.options[cryptoSuite.selectedIndex].value;
    if (cryptoSuiteName == CRYPTO_SUITE_PBKDF2_SHA256_AES_256_GCM.cryptoSuiteName) {
        return CRYPTO_SUITE_PBKDF2_SHA256_AES_256_GCM;
    } else if (cryptoSuiteName == CRYPTO_SUITE_PBKDF2_SHA256_AES_256_CBC.cryptoSuiteName) {
        return CRYPTO_SUITE_PBKDF2_SHA256_AES_256_CBC;
    } else {
        console.error("[-] Unable to find the rigth CryptoSuite");
        return;
    }
}

async function encryptBuffer(fileContentBuffer, cryptoSuite) {
    // STEP 01. Password devivation
    // 01.a. Import password
    var passwordKeyObj = await window.crypto.subtle.importKey(
        'raw',
        getPasswordFromInput(),
        {name: cryptoSuite.kdfAlgoName},
        false,
        ['deriveBits']
    )
    .catch(function(err) {
        console.error("[-] importKey KDF error: ", err);
    });
    console.log("[+] importKey KDF successful");
    // 01.b Derive password
    var saltBuffer = crypto.getRandomValues(new Uint8Array(8));
    var masterSecretObj = await window.crypto.subtle.deriveBits(
        {
            name: cryptoSuite.kdfAlgoName,
            salt: saltBuffer,
            iterations: cryptoSuite.kdfIterations,
            hash: {name: cryptoSuite.kdfHashFunctionName}
        },
        passwordKeyObj,
        384
    )
    .catch(function(err) {
        console.error("[-] deriveBits KDF error: ", err);
        errorMessage("Error encrypting file. See console log");
        return;
    });
    console.log("[+] deriveBits KDF successful");
    var masterSecretBytes = new Uint8Array(masterSecretObj)
    var encryptionKey = masterSecretBytes.slice(0, 32)
    var encryptionIv = masterSecretBytes.slice(32, 48)


    // STEP 02. Encrypt buffer
    var encryptionKeyObj = await window.crypto.subtle.importKey(
        'raw', encryptionKey,
        {
            name: cryptoSuite.encryptAlgoName,
            length: cryptoSuite.encryptKeyLengthBits
        },
        false,
        ['encrypt']
    )
    .catch(function(err){
        console.error("[-] importKey Encrypt error: ", err);
        errorMessage("Error encrypting file. See console log");
        return;
    });
    console.log("[+] importKey Encrypt successful");

    var cipherObj = await window.crypto.subtle.encrypt(
        {
            name: cryptoSuite.encryptAlgoName,
            iv: encryptionIv
        },
        encryptionKeyObj,
        fileContentBuffer
    )
    .catch(function(err){
        console.error("[-] encrypt error: ", err);
        errorMessage("Error encrypting file. See console log");
        return;
    });
    if(!cipherObj) {
        errorMessage("Error encrypting file. See console log");
        return;
    }
    console.log("[+] encrypt Encrypt successful");

    // STEP 03. Build encrypted buffer
    cipherBuffer = new Uint8Array(cipherObj);
    if(openssl.checked) {
        console.log("[+] OpenSSL compatible");
        encryptedBufferLengthBytes = cipherBuffer.length + 16;
        encryptedBuffer = new Uint8Array(encryptedBufferLengthBytes)
        encryptedBuffer.set(new TextEncoder("utf-8").encode('Salted__'));
        encryptedBuffer.set(saltBuffer, 8);
        encryptedBuffer.set(cipherBuffer, 16);
    } else {
        console.log("[+] OpenSSL not compatible");
        encryptedBufferLengthBytes = cipherBuffer.length + 8;
        encryptedBuffer = new Uint8Array(encryptedBufferLengthBytes)
        encryptedBuffer.set(saltBuffer);
        encryptedBuffer.set(cipherBuffer, 8);
    }
    successMessage("File encrypted!");
    console.log("[+] File encrypted");

    return encryptedBuffer;
}

/*********** File/Network functions *********/
function readfile(file){
    return new Promise((resolve, reject) => {
        var fr = new FileReader();
        fr.onload = () => {
            resolve(fr.result )
        };
        fr.readAsArrayBuffer(file);
    });
}


async function encryptThenUpload() {
    if (inputFile.value && password.value) {
        // 01. Encrypt the file
        const file = inputFile.files[0];
        var fileContent = await readfile(file)
            .catch(function (err) {
                console.error(err);
            });
        var plainBuffer = new Uint8Array(fileContent);
        var encryptedBuffer = await encryptBuffer(plainBuffer, getCryptoSuite());

        // 02. Send the encrypted buffer
        var formData = new FormData();
        formData.append("uploadFile", new Blob([encryptedBuffer]), file.name);
        formData.append("openssl", false);
        formData.append("", false);
        uploadFromData('/upload', formData);

    } else {
        console.log("[-] Please specify a file and enter a password");
        errorMessage("Please specify a file and enter a password");
        return;
    }
}

function uploadFromData(url, formData) {
    fetch(url, {
        method: 'POST',
        mode: 'cors', // no-cors, *cors, same-origin
        cache: 'no-cache',
        credentials: 'same-origin', // include, *same-origin, omit
        //headers: {
        //    'Content-Type': 'multipart/form-data'
        //},
        redirect: 'follow', // manual, *follow, error
        referrerPolicy: 'no-referrer', // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
        body: formData
    })
        .then(function(response) {
            console.log(response.status);
            console.log(response.ok);
            return response.json();
        })
        .then(function(result) {
            console.log(JSON.stringify(result));
            uploadFileStatusMessage(result.code, result.path, result.sha256sum);
        })
        .catch(function(err) {
            console.error("There has been a problem with your fetch operation: ", err);
            errorMessage("Upload file failed");
        })
    ;
}

/*********** HTML code related *********/
function errorMessage(message) {
    var errTag =
        `<div class="alert alert-danger alert-error alert-fade" role="alert">
        <button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button>
        <strong>Error:</strong> ${message}
        </div>`;
    document.getElementById("dynamicMessages").insertAdjacentHTML('beforeEnd', errTag);
    window.setTimeout(function () {
        $(".alert-fade").fadeTo(500, 0).slideUp(500, function () {
            $(this).remove();
        });
    }, 4000);
}

function successMessage(message) {
    var successTag =
        `<div class="alert alert-success alert-error alert-fade" role="alert">
        <button type="button" class="close" data-dismiss="alert" aria-label="Close"><span aria-hidden="true">&times;</span></button>
        <strong>Success:</strong> ${message}
        </div>`;
    document.getElementById("dynamicMessages").insertAdjacentHTML('beforeEnd', successTag);
    window.setTimeout(function () {
        $(".alert-fade").fadeTo(500, 0).slideUp(500, function () {
            $(this).remove();
        });
    }, 4000);
}

function uploadFileStatusMessage(code, path, sha256sum) {
    var htmlTag = `
<div class="alert alert-success alert-error" role="alert">
    <pre>
code: ${code}
path: ${path}
sha256: ${sha256sum}
    </pre>
</div>
`;

    uploadFileStatus.insertAdjacentHTML('afterbegin', htmlTag);
}

function showResetButton() {
    resetButton.style.display = "";
}

function hideResetButton() {
    resetButton.style.display = "none";
}

function updateNameAndSize() {
    showResetButton();
    let nBytes = 0,
        oFiles = inputFile.files,
        nFiles = oFiles.length,
        placeHolder = document.getElementById("file-placeholder");

    for (let nFileId = 0; nFileId < nFiles; nFileId++) {
        nBytes += oFiles[nFileId].size;
        fileName = oFiles[nFileId].name;
    }
    let sOutput = nBytes + " bytes";
    // multiples approximation
    for (let aMultiples = ["KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"], nMultiple = 0, nApprox = nBytes / 1024; nApprox > 1; nApprox /= 1024, nMultiple++) {
        sOutput = nApprox.toFixed(2) + " " + aMultiples[nMultiple];
    }
    // console.log(fileName);
    //change placeholder text
    if (!inputFile.value) {
        placeHolder.innerHTML = "Choose a file to encrypt/decrypt";
    } else {
        placeHolder.innerHTML = fileName + '  <span class="text-success">' + sOutput + '</span>';
    }
}

function resetInputs(){
    inputFile.value = "";
    password.value = "";
    updateNameAndSize();
    hideResetButton();
}