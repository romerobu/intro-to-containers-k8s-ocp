# Introduction to OpenShift

In this lab we are going to see how a developer and an admin can access the OpenShift Console in order to get an application deployed.

##  Configuring the Classroom Environment

Go to section 1.7 and follow the instructions to setup the configuration for your lab.

* You must have a github and quay account. You will be required to fork a project on your github account however for quay you just need a valid user.

## Create application using the Web Console

Go to section 6.10 and follow the instructions to create your first app with the web console.

* Some useful commands:
```bash
oc whoami --show-console # To retrieve console url
oc whoami --show-server # To retrieve OCP API
oc whaomi # To retrieve user logged
```

* Once you log in to the Openshift console for the first time is highly recommended to get started with the tour.

* On Developer Catalog, select PHP (Builder Images). We will use Templates on the next exercise.

## Create application using Templates

Go to section 7.6 to implement an app using Templates.

## Admin Console

The instructor will show:

- `Overview`
- `Operators`
- `Observe`
- `Compute`
- `User Management`
- `Administration - Cluster Settings`
