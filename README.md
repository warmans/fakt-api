# Stressfaktor API

Parses stressfaktor termine pages and provides data in a structured format.

## Endpoints

### `GET /api/v1/event`

| Param         | Description    | Example Value   | Notes                              |
| ------------- | -------------- | --------------- | ---------------------------------- |
| __event__     | event ID(s)    | 101             | or multiple comma seperated values |
| __from__      | event date     | 2012-11-10      | Y-m-d                              |
| __to__        | event date     | 2012-11-10      | Y-m-d                              |
| __venue__     | venue ID(s)    | 123             |                                    |
| __type__      | event type     | Konzert         |                                    |
| __deleted__   | show deleted   | 1               | show events in the past            |


#### Examples

`/api/v1/event?event=1,2,3` - show events with IDs 1, 2 or 3

`/api/v1/event?from=2016-01-01&to=2017-01-01` - show events in 2016


## Installing

Deb packages are provided in the dist directory. This is the easiest way to install. If you have a Go environment
you can build different packages via the `make package` target. e.g. `PACKAGE_TYPE=rpm make package`.


## Running

The deb includes a init script:

    /etc/init.d/stressfaktor-api start

or you can just run the binary some other way:

```
Usage of stressfaktor-api:
   -bind string
     	Web server bind address (default ":8080")
   -dbpath string
     	Location of DB file (default "./db.sqlite3")
   -location string
     	Time localization (default "Europe/Berlin")
   -termin string
     	Address of termine page (default "https://stressfaktor.squat.net/termine.php?display=90")
   -v	Print version and exit
```

## Building

`make build` or ` go get && go build`
