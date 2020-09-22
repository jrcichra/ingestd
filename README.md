# ingestd
+ HTTP server that easily ingests data into a database (powered by [gin](https://github.com/gin-gonic/gin)!)
+ Just POST JSON to http://hostname:port/database/table

# Usage
+ Specify your database credentials in `config.txt`
    ```
    username:password@tcp(hostname)/
    ```
+ Run the container, mapping in the config file with the credentials:

    ```
    docker run -it -d \
    --name=ingestd \
    --hostname=ingestd \
    --restart=unless-stopped \
    -v ${PWD}/config.txt:/ingestd/config.txt \
    -p 8080:8080 \      # Change left side port to your liking
    jrcichra/ingestd
    ```
+ Try simple GET to make sure gin is up:
    ```
    $ curl http://hostname:port/ping
    {"message":"pong"}
    $
    ```
+ Make a POST call (i.e cURL) - see [post.sh](post.sh):
    ```
    curl -i --header "Content-Type: application/json" \
    --request POST \
    --data "{"col1": "value1", "col2": "value2"}" \
    http://hostname:port/database/table
    ```
    Form the URL and JSON payload to match your database schema.

    If the insert was successful, the HTTP server will return an empty 200 OK response. Any other issue will return a 500 with a JSON body of the server-side error.