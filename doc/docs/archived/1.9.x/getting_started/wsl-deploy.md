# Deploy Pachyderm on a Microsoft Windows machine

You can deploy Pachyderm locally on your Microsoft Windows computer to
try out Pachyderm. However, currently Pachyderm has limited support for
Windows. This document provides guidelines that enable you to run Pachyderm
on a Windows machine for preview. You might need to troubleshoot these
instructions further to make this work. To test full functionality, we
encourage you to run Pachyderm in a UNIX environment.

## Prerequisites

You need to have the following components installed on your Microsoft Windows
computer to deploy Pachyderm:

* Microsoft Windows 10 or later
* Microsoft PowerShell™
* Ubuntu™ Windows Subsystem for Linux (WSL)
* Hyper-V™
* Minikube
* `kubectl`

While PowerShell is included with your standard Windows tools, other components
might not be included or enabled. Pachyderm recommends that you install Ubuntu
WSL because it was tested and verified by our team.

## Install Ubuntu WSL

`pachctl` is a Pachyderm command-line interface that enables you to interact
with Pachyderm. It requires a UNIX environment to run correctly. Therefore,
you need a WSL to run `pachctl` commands. We recommend that you install an
Ubuntu WSL. To install WSL, follow the steps in the
[Microsoft documentation](https://docs.microsoft.com/en-us/windows/wsl/install-win10).

## Enable Hyper-V

You need to install Hyper-V on your host Windows machine so that Minikube can
use it to create a virtual machine on which your Pachyderm containerized cluster
will run. You also need to adjust the amount of CPU and memory allocated to
Hyper-V so that your Minikube machines do not crash with an out-of-memory error.

To enable Hyper-V, complete the following steps:

1. Run PowerShell as Administrator.
1. Enable Hyper-V by running the following command:

   ```powershell
   Enable-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V -All
   ```

1. Adjust resource utilization by opening the Hyper-V manager. Typically,
setting memory to 3072 MB should be sufficient for most testing deployments.

## Install Minikube

Minikube is a tool that enables you to run a single-node Kubernetes cluster that
is typically used for testing locally on your computer.
You need to install Minikube on your host machine from PowerShell and not from
the Ubuntu WSL. In this section, Minikube is installed by using the Chocolatey
package manager, which is a popular tool for installing programs on Windows. But if
you prefer another installation method, it should work too.

To install Minikube, complete the following steps:

1. Install Chocolatey as described in the [Chocolatey documentation](https://chocolatey.org/docs/installation).
1. Install Minikube:

   ```shell
   choco install minikube
   ```

## Install `kubectl`

You need to install `kubectl` on your Ubuntu WSL and then copy the `kubectl`
configuration file from MInikube that is installed on your host Windows
machine to the `~/.kube` directory in your Ubuntu WSL machine.

To install and configure `kubectl`

1. Install `kubectl` on WSL as described in the
[Kubernetes documentation](https://kubernetes.io/docs/tasks/tools/install-kubectl/).
1. Create a `~/.kube` directory.
1. Copy the Minikube config file from the Windows host system to the Ubuntu
WSL `/.kube/` directory:

   ```shell
   $ cp /mnt/c/Users/Svetlana/.kube/config ~/.kube/config
   ```

1. Change the following lines in the configuration file:

   ```shell hl_lines="3 4 5"
    users:
    - name: minikube
      user:
         client-certificate: C:\Users\User\.minikube\client.crt
         client-key: C:\Users\User\.minikube\client.key
   ```

   You need to change the path to the client key and certificate
   to point them to the location on your Windows host.

1. Verify that `kubectl` is configured correctly:

   ```shell
   $ kubectl version
   Client Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.0", GitCommit:"70132b0f130acc0bed193d9ba59dd186f0e634cf", GitTreeState:"clean", BuildDate:"2019-12-07T21:20:10Z", GoVersion:"go1.13.4", Compiler:"gc", Platform:"linux/amd64"}
   Server Version: version.Info{Major:"1", Minor:"17", GitVersion:"v1.17.0", GitCommit:"70132b0f130acc0bed193d9ba59dd186f0e634cf", GitTreeState:"clean", BuildDate:"2019-12-07T21:12:17Z", GoVersion:"go1.13.4", Compiler:"gc", Platform:"linux/amd64"}
   ```

## Install `pachctl`

Get the latest version of `pachctl` by directly downloading it to your computer
as described in [Install pachctl](../local_installation/#install-pachctl).

**Example:**

```shell
$ curl -o /tmp/pachctl.deb -L https://github.com/pachyderm/pachyderm/releases/download/v1.10.0/pachctl_1.10.0_amd64.deb && sudo dpkg -i /tmp/pachctl.deb
% Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   613    0   613    0     0   2043      0 --:--:-- --:--:-- --:--:--  2043
100 25.9M  100 25.9M    0     0  10.0M      0  0:00:02  0:00:02 --:--:-- 13.0M
```

## Deploy Pachyderm

After you complete all the steps above, start your Minikube VM by running
`minikube start` and deploy Pachyderm in Ubuntu WSL by running
`pachctl deploy local` as described in [Deploy Pachyderm](../local_installation/#deploy-pachyderm).
