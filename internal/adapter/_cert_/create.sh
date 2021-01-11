#!/bin/bash

SSL_EXPIRE_DAYS=10950
SSL_KEY_BITS=2048

SSL_C="CN"                  # Country Name (2 letter code)
SSL_ST="X Province"         # State or Province Name (full name)
SSL_L="X City"              # Locality Name (eg, city)
SSL_O="X Company"           # Organization Name (eg, company)
SSL_OU="IT"                 # Organizational Unit Name (eg, section)
SSL_CN="www.example.com"    # Common Name (eg, fully qualified host name)

readonly CURRENT_DIR=$(cd $(dirname "${BASH_SOURCE[0]}") && pwd)


echo $CURRENT_DIR

# ${1} directory full path
function makeEmptyDirectoryAndEnter() {
  if [ -d "${1}" ]; then
      rm -rf ${1}
  fi

  mkdir -p ${1}

  cd ${1}
}

# ${1} Output directory
function initCA() {
  if [ ! -d "${1}" ]; then
    makeEmptyDirectoryAndEnter ${1}
    cat > ca.cnf <<EOF
[ req ]
 default_bits           = ${SSL_KEY_BITS}
 distinguished_name     = req_distinguished_name
 prompt                 = no
[ req_distinguished_name ]
 C                      = ${SSL_C}
 ST                     = ${SSL_ST}
 L                      = ${SSL_L}
 O                      = ${SSL_O}
 OU                     = ${SSL_OU}
 CN                     = ${SSL_CN}
[ v3_ca ]
keyUsage = critical, keyCertSign, cRLSign
basicConstraints = critical, CA:TRUE, pathlen:2
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
EOF
    openssl genrsa -out ca-key.pem ${SSL_KEY_BITS}
    openssl req -x509 -new -nodes -config ca.cnf -extensions v3_ca -key ca-key.pem -days ${SSL_EXPIRE_DAYS} -out ca.pem
    removeFile ca.cnf
  fi
}


# ${1} CA directory
# ${2} Output directory
# ${3} CN (common name)
# ${4} ip array string
# ${5} dns array string
function makeSSLAuthorFiles() {
  makeEmptyDirectoryAndEnter ${2}

  local embedString=""
  local index=1
  local ip
  local first="yes"

  for ip in ${4}; do
    if [ "${first}" == "yes" ];then
        first="no"
    else
        embedString=${embedString}$'\n'
    fi

    embedString=${embedString}"IP.${index} = ${ip}"
    ((index+=1))
  done

  index=1
  for dns in ${5}; do
    if [ "${first}" == "yes" ];then
        first="no"
    else
        embedString=${embedString}$'\n'
    fi

    embedString=${embedString}"DNS.${index} = ${dns}"
    ((index+=1))
  done

  if [ "${embedString}" == "" ]; then
    cat > ${3}.cnf <<EOF
[ req ]
 default_bits           = ${SSL_KEY_BITS}
 distinguished_name     = req_distinguished_name
 prompt                 = no
[ req_distinguished_name ]
 C                      = ${SSL_C}
 ST                     = ${SSL_ST}
 L                      = ${SSL_L}
 O                      = ${SSL_O}
 OU                     = ${SSL_OU}
 CN                     = ${3}
[ v3_req ]
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
EOF
  else
    cat > ${3}.cnf <<EOF
[ req ]
 default_bits           = ${SSL_KEY_BITS}
 distinguished_name     = req_distinguished_name
 prompt                 = no
[ req_distinguished_name ]
 C                      = ${SSL_C}
 ST                     = ${SSL_ST}
 L                      = ${SSL_L}
 O                      = ${SSL_O}
 OU                     = ${SSL_OU}
 CN                     = ${3}
[ v3_req ]
keyUsage = critical, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth, clientAuth
basicConstraints = critical, CA:FALSE
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
subjectAltName = @alt_names
[alt_names]
${embedString}
EOF
  fi

  openssl genrsa -out ${3}-key.pem ${SSL_KEY_BITS}
  openssl req -new -key ${3}-key.pem -out ${3}.csr -config ${3}.cnf
  openssl x509 -req -in ${3}.csr -CA ${1}/ca.pem -CAkey ${1}/ca-key.pem -CAserial ca.srl \
    -CAcreateserial -out ${3}.pem -days ${SSL_EXPIRE_DAYS} -extensions v3_req -extfile ${3}.cnf
}

initCA ${CURRENT_DIR}/ca
makeSSLAuthorFiles ${CURRENT_DIR}/ca ${CURRENT_DIR}/server server "127.0.0.1 ::1" "localhost"
makeSSLAuthorFiles ${CURRENT_DIR}/ca ${CURRENT_DIR}/client client "" ""
