import os
import sys

from Crypto.Cipher import AES
from Crypto.Hash import SHA256
from Crypto.Protocol.KDF import PBKDF2
from Crypto.Util.Padding import unpad


SALT_SZ_IN_BYTES = 8
MASTER_KEY_SZ_IN_BYTES = 48
AES_GCM_TAG_SZ_IN_BYTES = 16
CRYPTO_SUITES = ['pbkdf2-sha256-aes-256-gcm', 'pbkdf2-sha256-aes-256-cbc']

def decrypt_data(crypto_suite, password, data):
    if 'pbkdf2-sha256-aes-256-gcm' == crypto_suite:
        try:
            # deriv password
            salt = data[:SALT_SZ_IN_BYTES]
            km = PBKDF2(password, salt, 48, count=32000, hmac_hash_module=SHA256)
            # decrypt buffer
            aes_key = km[:32]
            aes_gcm_nonce = km[32:48]
            aes = AES.new(aes_key, AES.MODE_GCM, nonce=aes_gcm_nonce)

            return aes.decrypt_and_verify(data[SALT_SZ_IN_BYTES:-AES_GCM_TAG_SZ_IN_BYTES], data[-AES_GCM_TAG_SZ_IN_BYTES:])
        except (ValueError, KeyError) as e:
            return None
    elif 'pbkdf2-sha256-aes-256-cbc' == crypto_suite:
        try:
            # deriv password
            salt = data[:SALT_SZ_IN_BYTES]
            km = PBKDF2(password, salt, 48, count=32000, hmac_hash_module=SHA256)
            # decrypt buffer
            aes_key = km[:32]
            aes_iv = km[32:48]
            aes = AES.new(aes_key, AES.MODE_CBC, aes_iv)

            return unpad(aes.decrypt(data[SALT_SZ_IN_BYTES:]), AES.block_size)
        except (ValueError, KeyError) as e:
            return None
    else:
        return None

def print_usage(argv):
    print("usage: {0} <crypto_suite> <password> <input_file> <output_file>".format(argv[0]))
    print('  crypto suites available:')
    for cs in CRYPTO_SUITES:
        print('    {}'.format(cs))

def main(argv):
    # 00. check arguments
    if len(argv) != 5:
        print_usage(argv)
        sys.exit(1)
    # end if
    _crypto_suite = argv[1]
    _password = argv[2]
    _input_file_path = argv[3]
    _output_file_path = argv[4]

    if _crypto_suite not in CRYPTO_SUITES:
        print_usage(argv)
        sys.exit(1)

    with open(_input_file_path, 'rb') as fin, open(_output_file_path, 'wb') as fout:
        data = fin.read()
        decrypted_data = decrypt_data(_crypto_suite, _password, data)
        if decrypted_data is not None:
            fout.write(decrypted_data)


if __name__ == "__main__":
    main(sys.argv)