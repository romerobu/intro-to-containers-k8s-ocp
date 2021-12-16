# Introduction to Containers

In this lab we are going to see how we can use Podman to run containers in our local system.

## Lab 1 - Running our first containers

In this lab we will run some containers and learn how we can expose services running in a container to our local system.

1. Connect to the Fedora 35 system
2. Run a hello-world container

    ~~~sh
    podman run docker.io/library/hello-world:latest
    ~~~
3. Containers can run in background, let's run an Apache server on background

    > **NOTE**: `-d` flag will detach the container (run in background), `--rm` flag will remove the container once it stops running.
    
    ~~~sh
    podman run -d --rm docker.io/library/httpd:2.4
    ~~~
4. We don't see any output, just the container ID, but we can verify that the container is running and also get the container logs:

    1. Get running containers
        
        ~~~sh
        podman ps

        CONTAINER ID  IMAGE                        COMMAND           CREATED         STATUS             PORTS       NAMES
        b39be8e87a0b  docker.io/library/httpd:2.4  httpd-foreground  43 seconds ago  Up 43 seconds ago              magical_khorana
        ~~~
    2. Get container logs
    
        > **NOTE**: The container got a random name assigned `magical_khorana` in the example above, you can specify a name when running the container that will make things easier when trying to interact with a specific container.

        ~~~sh
        # We can use the container id or the container name
        podman logs b39be8e87a0b

        AH00558: httpd: Could not reliably determine the server's fully qualified domain name, using 10.88.0.3. Set the 'ServerName' directive globally to suppress this message
        AH00558: httpd: Could not reliably determine the server's fully qualified domain name, using 10.88.0.3. Set the 'ServerName' directive globally to suppress this message
        [Thu Dec 16 11:19:37.618618 2021] [mpm_event:notice] [pid 1:tid 140512215055680] AH00489: Apache/2.4.51 (Unix) configured -- resuming normal operations
        [Thu Dec 16 11:19:37.618745 2021] [core:notice] [pid 1:tid 140512215055680] AH00094: Command line: 'httpd -D FOREGROUND'
        ~~~
5. We have an apache running, but how do we access it? - Well, we need to expose ports for that, let's stop the previous container and create a new one that exposes container port 80.

    1. Stop the previous container
        
        ~~~sh
        podman kill b39be8e87a0b
        ~~~
    2. Run the new container

        > **NOTE**: `-p` flag is used to expose container ports to the local node. In this case we're exposing container's port 80 in host's port 8080

        ~~~sh
        podman run -d --rm --name apache-container -p 8080:80 docker.io/library/httpd:2.4
        ~~~
    3. If we open our browser or we curl our node IP on port 8080 we will get to the apache server:

        ~~~sh
        curl http://127.0.0.1:8080/

        <html><body><h1>It works!</h1></body></html>
        ~~~
    4. We can stop the container

        ~~~sh
        podman kill apache-container
        ~~~

## Lab 2 - Exploring a container

In this lab we will connect to a running container and check the processes running and its filesystem.

1. Run a container based in ubuntu and change the entrypoint to `sleep infinity`

    > **NOTE**: Container images can define entrypoints that will be executed by default, in the example below we're overwritting the default entrypoint so the container runs the sleep command instead.
    
    ~~~sh
    podman run -d --rm --entrypoint '["sleep", "infinity"]' --name ubuntu-container docker.io/library/ubuntu:22.04
    ~~~
2. Let's connect to the container by running a shell inside it

    > **NOTE**: `-t` flag attaches a pseudo-tty and `-i` flag keeps stdin open.

    ~~~sh
    podman exec -ti ubuntu-container /bin/bash
    ~~~
3. We're inside the container now, let's take a look at the file system, for example let's cat the os-release file:

    ~~~sh
    cat /etc/os-release

    PRETTY_NAME="Ubuntu Jammy Jellyfish (development branch)"
    NAME="Ubuntu"
    VERSION_ID="22.04"
    <output_omitted>
    ~~~
4. As you can see the container filesystem is different than the host filesystem. What about processes, let's see how many processes are running in the container:

    ~~~sh
    ps -ef

    UID          PID    PPID  C STIME TTY          TIME CMD
    root           1       0  0 11:55 ?        00:00:00 sleep infinity
    root           4       0  0 11:57 pts/0    00:00:00 /bin/bash
    root          10       4  0 11:59 pts/0    00:00:00 ps -ef
    ~~~
5. As we mentioned during the presentation, the container cannot see the processes running on the host, it only sees its own processes. We can exit the container:

    ~~~sh
    exit
    ~~~
6. We can stop the container now

    ~~~sh
    podman kill ubuntu-container
    ~~~
7. In the previous example we ran `sleep` as entrypoint but we can run other programs such a `bash` shell as well.

## Lab 3 - Building your very first container image

In the previous labs we have been using different images, in this case we are going to build our own container image, it will be a simple one.

We will be using the [Dockerfile format](https://docs.docker.com/engine/reference/builder/#format). In order to build and run our [test application](https://github.com/mvazquezc/reverse-words).

1. Create the following Containerfile in your Fedora35 system

    ~~~sh
    cat <<EOF > /var/tmp/reversewords-containerfile
    # Use golang:1.16 as base image since our app is programmed in go and targets this go release
    FROM docker.io/library/golang:1.16
    # Set our working directory to /tmp/
    WORKDIR /tmp/
    # Clone the git repository with the application's code
    RUN git clone https://github.com/mvazquezc/reverse-words/
    # Access the repository folder, download dependencies and build the app
    RUN cd reverse-words && go mod tidy && \
        CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reverse-words . && \
        cp reverse-words /usr/bin/reverse-words && \
        chmod +x /usr/bin/reverse-words
    # App executes as user 9999 since it doesn't need root privileges or uid
    USER 9999
    # App exposes port 8080
    EXPOSE 8080
    # Default command will be our app
    CMD ["/usr/bin/reverse-words"]
    EOF
    ~~~
2. We can use podman to build our container image

    > **NOTE**: After a few moments we will get the image tagged locally as `localhost/reverse-words:latest`.

    ~~~sh
    podman build -f /var/tmp/reversewords-containerfile -t reverse-words:latest

    <omitted_output>
    Successfully tagged localhost/reverse-words:latest
    ~~~
3. We can now run and access our application

    ~~~sh
    podman run -d --rm --name reverse-words -p 8080:8080 localhost/reverse-words:latest
    ~~~
4. Our application exposes an API that reverses words, let's try it:

    ~~~sh
    curl http://127.0.0.1:8080/ -X POST -d '{"word": "Hello!"}'

    {"reverse_word":"!olleH"}
    ~~~
5. We can stop the container now:

    ~~~sh
    podman kill reverse-words
    ~~~
6. After testing our application we could push the container image to a container registry such as dockerhub or quay.io, in order to do that we need to tag our local image:

    > **NOTE**: Below example would tag our image so we can push it into quay.io under the `mavazque` account:

    ~~~sh
    podman tag localhost/reverse-words:latest quay.io/mavazque/reverse-words:latest
    ~~~
7. Now that the image is tagged we can check that we have that image tag localy:

    ~~~sh
    podman images 

    REPOSITORY                      TAG         IMAGE ID      CREATED         SIZE
    quay.io/mavazque/reverse-words  latest      099d96771a3c  10 minutes ago  1.04 GB
    <omitted_output>
    ~~~
8. We can now push the image:

    > **NOTE**: This step will fail since we are not authenticated in quay.io

    ~~~sh
    podman push quay.io/mavazque/reverse-words:latest

    <omitted_output>
    unauthorized: authentication required
    ~~~

## Lab 4 - Docker-compose, Let's run Pacman!

If you remember, we talked about multi-container applications. In this case we are going to deploy a Pacman game that uses MongoDB to store player's highscores.

1. Create the following docker-compose in your Fedora35 system

    ~~~sh
    cat <<EOF > /var/tmp/pacman-app.yaml
    version: '3'
    services:
      pacman:
        image: "quay.io/ifont/pacman-nodejs-app:latest"
        depends_on:
          - "mongo"
        links:
          - mongo
        restart: always
        hostname: pacman
        environment:
          MONGO_SERVICE_HOST: mongo
          MONGO_AUTH_USER: user
          MONGO_AUTH_PWD: password
          MONGO_DATABASE: admin
          MY_MONGO_PORT: 27017
          MY_NODE_NAME: pacman
        ports:
          - "8080:8080"
      mongo:
        image: "docker.io/library/mongo:latest"
        hostname: mongo
        restart: always
        environment:
          MONGO_INITDB_ROOT_USERNAME: user
          MONGO_INITDB_ROOT_PASSWORD: password
    EOF
    ~~~
2. We can start our Pacman app by running podman-compose that will read the application definition we created in the previous steps and will create the required containers.

    > **NOTE**: `-f` flag is used to specify a locations for the compose manifest. `-d` flag is used to detach from the container (run in background)

    ~~~sh
    podman-compose -f /var/tmp/pacman-app.yaml up -d
    ~~~
3. We can access our Pacman game in our local server in port 8080. Open your favorite web-browser and go to [http://127.0.0.1:8080](http://127.0.0.1:8080).
4. Once we're done playing our game we can power down the application stack:

    ~~~sh
    podman-compose -f /var/tmp/pacman-app.yaml down
    ~~~
