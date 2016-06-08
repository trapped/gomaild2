@smtp
Feature: Connection
  A client should be able to connect to the server.

  Scenario: Client connects and is greeted through TCP
    Given a server is listening on "127.0.0.1:2525"
    When a client connects to "127.0.0.1:2525"
    Then the client should receive a 220 reply
