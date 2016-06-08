@smtp
Feature: Identification
  A client connected to the server should be able to identify.

  Scenario:
    Given a server is listening on "127.0.0.1:2525"
    And a client is connected to "127.0.0.1:2525"
    When the client sends a "HELO" command with args <domain>
    Then the client should receive a <code> reply

    Examples:
      | domain        | code |
      | "localhost"   | 250  |
      | "."           | 501  |
      | "aaaa"        | 501  |
      | "example.com" | 250  |
