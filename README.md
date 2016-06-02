gomaild2
--------
Hopefully the successor to [`gomaild`](https://github.com/trapped/gomaild).

Requires [`gengen`](https://github.com/trapped/gengen) to generate some files (`go get github.com/trapped/gengen`).

##What can it do?
With `gomaild2` you can:

- send and receive email to/from most other mail servers/service providers
- force SSL/TLS encryption during mail exchange (when sending)
- safely authenticate (encrypting your credentials) when fetching email with POP3
- setup a catch-all email server for verifying throwaway accounts
- easily backup the database

##Setting up
Most people only need to receive and send simple email, so we'll go over how to set `gomaild2` up for that.

###Building
Right now the only way to get `gomaild2` is to build from source (but we plan to change that in the near future).
To do that, you're gonna need the latest version of the Go programming language and set it up. I'm going to assume you can search yourself how to do that.

Next, you need to download `gomaild2`'s source code:

`go get github.com/trapped/gomaild2`

If the previous command didn't install `gomaild2` for you (check if `$GOPATH/bin/gomaild2` exists) move into the source directory and build again:

```bash
$ cd $GOPATH/src/github.com/trapped/gomaild2
$ go install
```

###Preparing the configuration file
Great, so now you have a working `gomaild2` binary. What now? Well, you obviously need a configuration file.
We provide a "default" one, which should be fine for most purposes, but it's mostly a placeholder, so we'll go over creating a config file from scratch. You can start by creating a `config.yaml` file in your home directory:

```bash
$ cd
$ touch config.yaml
```

####The boring stuff
First off you need to setup things like logging and server settings. Copy-paste this into the config file and leave everything as-is except the server name (unless you know what you're doing):

```yaml
log:
  path: gomaild2.log

server:
  name: YOUR DOMAIN NAME
  smtp:
    mta:
      address: 0.0.0.0
      ports:
        - 25
      timeout: 600 # seconds
    msa:
      address: 0.0.0.0
      require_auth: true
      outbound: true
      ports:
        - 587
      timeout: 600 # seconds
  pop3:
    local:
      address: 0.0.0.0
      ports:
        - 110
      timeout: 600 # seconds
```

A little note about `server.name`: it should be the domain name that points to the machine with the email server on.

Say for example you had a VPS, and say you set up `mail.foobar.com` to point to its IP. Then `server.name` should be `mail.foobar.com`, although usually you can also just use `foobar.com`.

####The transfer agent
The transfer agent (and its workers) is the small piece of code that takes the email you send and makes sure they arrive at destination. You can tweak a couple settings:

```yaml
transfer:
  max_tries: 3
  worker_count: 1
  allow_unencrypted: true
  allow_insecure: true
```

`max_tries` is the amount of times workers will try to connect to the receiving email server. 3 times is usually fine.

`worker_count` is the amount of workers that will be available to transfer your email. If you plan on having many users and/or sending many emails very fast, you can increment this.

When `allow_unencrypted` is set to false, workers will refuse to transfer your email to servers that don't support SSL/TLS encryption. Generally they support it, but you never know.

`allow_insecure` makes workers accept to transfer your email to servers that don't have a 'valid' SSL/TLS certificate, such as those using a self-signed one.

####The database
```yaml
db:
  save_all_mail: false
  path: gomaild2.db
```

`save_all_mail` makes your server accept and save email belonging to addresses that don't exist (yet). If you add an account later, you'll be able to read even emails received before you created it.

`path` is the place where the database file is. You might find it useful when backing it up.

####Encryption and accounts
It is never a good idea to let strangers read what you write. Hence why you can encrypt your email while in transit, as well as the credentials for your email accounts!

To encrypt credentials, first you need an AES256 key, which is any kind of data that is 32 bytes long (for example `123456789012345678901234567890ab`). After you have one, encode it to base64 (`echo -n "123456789012345678901234567890ab" | openssl base64`) and set the `PW_ENCRYPTION` environment variable to it, then actually encrypt your password (`echo -n "YOUR PASSWORD HERE" | openssl aes-256-cbc -a` then type the original non-base64 password) and set the base64-encoded result in the config file.

Encrypting credentials can be complicated, though, so if you want you can disable it by NOT setting `PW_ENCRYPTION`.

```yaml
domains:
  YOUR DOMAIN HERE:
    users:
      - YOUR USERNAME HERE@YOUR (MAYBE ENCRYPTED) PASSWORD HERE
      - ANOTHER USERNAME@ANOTHER PASSWORD
```

Another thing that is important to set up is SSL/TLS encryption. You first need a valid certificate (search on the internet how to get one, but I recommend [Let's Encrypt](https://letsencrypt.org)).

```yaml
tls:
  enabled: true
  certificate: /certificates/YOUR DOMAIN.crt
  key: /certificates/YOUR DOMAIN.key
```

###Running the email server
It would be a good idea to have a so-called supervisor 'babysit' `gomaild2` and restart it in case of crashes, as well as alert you if something goes wrong.
The easy (and dangerous!) way, though, is to simply run it while into the folder containing the configuration file: `$GOPATH/bin/gomaild2`

You will see lots of text, lots of log lines, but you can ignore it. You will also notice you can't type into your terminal anymore. If you're not using a supervisor, you can just install `screen` and make a 'virtual' terminal for the server:

```bash
$ screen -S gomaild2
... a blank screen opens
$ $GOPATH/bin/gomaild2
...
```

You'll see log lines again, but this time you can press `CTRL + A + D` to detach from the virtual screen. To check on it, you can either reattach (`screen -R gomaild2`) or read the log file (`cat gomaild2.log`). Some people also like to monitor the log file in real time (`tail -f gomaild2.log`).

To stop it, kill it like you would with any other program: `CTRL + C` (or `pkill`, etc...).

###Setting up Thunderbird
This is the easy part, since Thunderbird has a great automatic wizard that does most of the work for you; however:

- when asked to input any kind of USERNAME, use the WHOLE email address (`user@domain.com`, NOT `user`)
- if you didn't enable/setup SSL/TLS encryption, disable it in Thunderbird too
- when asked about what port to connect to, use `110` for POP3 and `587` for SMTP (unless you changed them)

###Backing up and resetting the database
`gomaild2` uses [`boltdb`](https://github.com/boltdb/bolt) as the underlying database, which means that backing up is as simple as copying the database file itself (PROVIDED THE SERVER IS NOT RUNNING):

```bash
$ cp gomaild2.db gomaild2.db.backup-$(date -I)
```

To reset the database just delete the file (PROVIDED THE SERVER IS NOT RUNNING):

```bash
$ rm gomaild2.db
```

##Status

- [x] DB interface to actually store email
- [x] YAML config
- [x] Logging (logfmt)
- [ ] SMTP server
  - [x] `HELO`
  - [x] `EHLO` (implicit `PIPELINING` and `8BITMIME`)
  - [x] `NOOP`
  - [x] `MAIL FROM`
  - [x] `RCPT TO`
  - [x] `DATA`
  - [x] `QUIT`
  - [x] `AUTH`
  - [x] `STARTTLS`
  - [ ] `DSN` (Delivery Status Notifications)
- [x] SMTP transfer agent/client (to send outbound email)
- [ ] POP3 server
  - [ ] `APOP`
  - [ ] `DELE`
  - [x] `LIST`
  - [x] `NOOP`
  - [x] `PASS`
  - [x] `CAPA`
  - [x] `QUIT`
  - [x] `RETR`
  - [x] `STLS`
  - [ ] `TOP`
  - [ ] `RSET`
  - [x] `STAT`
  - [x] `UIDL`
  - [x] `USER`
  - [x] `AUTH`