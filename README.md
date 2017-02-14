# go-driver

This process parses a file and returns its AST. A request contains the content of the file being analyzed as a string.

**Request** message has the following structure:

```
{
    "action": "ParseAST"
    "language": <language> (string, optional)
    "language_version": <language version> (number, optional)
    "path": <file path> (string)
    "content": <content> (string)
}
```
**Response** message is:

```
{
    "status": <status> ("ok", "error", "fatal")
    "errors": [ <error message>, <error message>, ... ]
    "driver" <driver name and version> (string)
    "language": <language> (string)
    "language_version": <language version> (string)
    "ast": <AST> (object)
}
```