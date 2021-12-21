# Introduction to OpenShift

In this lab we are going to see how a developer and an admin can access the OpenShift Console in order to get an application deployed.

## Running Pacman from the Operator

1. Access the [OpenShift Console URL](https://red.ht/ieselgrao)
2. Login with `RHA` using the credentials shared during the session
3. If you get a prompt for a console tour press `Skip Tour`
4. Click on `Create a new project` and name it after yourself
5. On the left menu click on `+Add`
6. Under `Developer Catalog` click on `Operator Backed`
7. Click on `Pacman Game`
8. Click on `Create`
9. Give it a name and specify the number of replicas (defaults are great)
10. Click on Create
11. You will get presented with the application view, click on the item that starts with `pacman...`
12. On the right menu, click on the Service that starts with `pacman-`
13. Under Service address you will get a `Location`, feel free to copy this location on your browser targetting the `http` port you can see in `Service port mapping`. i.e: `http://ac31dde5d315d4f0b866629f4fe8c765-280921920.eu-central-1.elb.amazonaws.com:8080`

## Create application from Source

1. Loged in the OpenShift Console as a developer click on `+Add`
2. Under `Developer Catalog` click on `All services`
3. Searc for `Go` and click on the `Go` builder image
4. Click on `Create Application`
5. Select `1.14.7-ubi8` as `Builder Image version`
6. Add `https://github.com/mvazquezc/reverse-words.git` as the `Git Repo URL`
7. Under `General` name the application.
8. Under `Resources` make sure `Deployment` is selected
9. Under `Advanced options` make sure `Create a route to the Application` is selected
10. Click on `Create`
11. Back in the `Topology` view, click on the `Go` application
12. Under `Build` you will see the application being built, feel free to check the logs
13. Once the build is done, you can access your application using the `Route` under `Routes`

## Admin Console

The instructor will show:

- `Overview`
- `Operators`
- `Observe`
- `Compute`
- `User Management`
- `Administration - Cluster Settings`