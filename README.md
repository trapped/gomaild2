gomaild2
--------
Hopefully the successor to [gomaild](https://github.com/trapped/gomaild).
Requires [`gengen`](https://github.com/trapped/gengen) to generate some files (`go get github.com/trapped/gengen`).

##Status

- [x] DB interface to actually store emails
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
- [x] SMTP transfer agent/client (to send outbound emails)
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