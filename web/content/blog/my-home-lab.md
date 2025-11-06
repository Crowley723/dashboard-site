---
title: My Homelab Journey
date: November 6, 2025
description: A detailed journey through my homelab setup, from Apache2 and Traefik to Proxmox clustering and Kubernetes orchestration.
image: https://images.unsplash.com/photo-1639322537228-f710d846310a?w=800&h=450&fit=crop
---

## Why a Homelab?
When my sister was given a pc, she didn't have a need for her old desktop pc (actually my old pc). So I got a new ssd and put linux on it. I started out messing with linux, I created my first websites in plain HTML, CSS, JS, and PHP (LAMP stack) which ran on the server using [apache2].

It was always meant to be a learning experience for me. I was studying computer science in community college and wanted something I could mess with that had really tangible effect (rather than just the basic menu-driven terminal apps for class).

Since then, I have evolved to a multiserver setup, running an 8 node Kubes ([K3s]) cluster in my lab. Running various applications, including the LGTMP ([Loki], [Grafana], [Tempo], [Mimir], [Pyroscope]) open source observability stack, [Authelia] for single sign-on and identity, [ArgoCD] for GitOps, GitHub Actions [Scale sets] (for using k8s to self-host ephemeral GitHub runners), multiple [Cloud-native PostgreSQL] databases, [Longhorn] for distributed storage and redundancy, [Cert-manager] for managing TLS certificates and running my local PKI, and much more.

The goal of this blog post is to document and discuss my journey (as much as I can remember).

## My Journey

### Apache2

I've been running my homelab for almost 5 years. It started out just as a way to run a couple basic websites using [apache2], then I got to the point where I was trying to run applications using apache2 as the reverse proxy for containerized applications (docker).

As anyone who has used apache2 as a reverse proxy can tell you, it gets really tedious to manage a large number of ports in apache2, you end up with a lot of duplication in the configuration. At the point where I was trying to track which ports were being used by which service (and even writing a python program to handle that for me), I decided it was time for a switch.

Thus began my journey with [Traefik].

### Discovery of Traefik

Traefik is a reverse proxy especially suited for containerized applications due to its incredible ability to hook into the docker socket and automatically create routes for containers based on labels. So instead of creating a file (similar to apache2) where you define each service, port, and domain that every app runs on, you define that info on the actual container definition. This has the benefit of avoiding one really large configuration file, and keeping an application's configuration (network, container, and volumes) in the same place.

Traefik was a bit of a learning curve coming from apache2, I also briefly looked at using nginx but gave up on that really quickly. After the initial difficulty with getting traefik to provision wildcard tls certificates, it was smooth sailing.

I could create a new docker compose project, define ~5 labels, and have traefik automatically route that container when it started. At this point, I containerized my existing websites (LAMP stack) so that I could redeploy them in docker. They exist in this state today.

### Proxmox Virtualization

At this point, all my containers and applications were on a single bare metal linux instance (Rocky Linux 9), which was fine, but I was craving the ability to run other operating systems (like HomeAssistant). I started looking into hypervisors and the natural choice is Proxmox.

A friend of mine who was already using Proxmox, mentioned that its possible (and actually fairly simple) to install Proxmox on a new drive and pass the existing drive (with Rocky linux) through to a virtual machine. Eventually I built up the courage to do this and completed it.

At the end, I had a Proxmox server that was running my existing Rocky linux machine as a virtual machine with all the existing containers (around 40 at this point).

### Authelia

Unfortunately, I wasn't nearly as invested in the idea of documenting my journey back then as I am now. So its hard to describe the state of my home lab and myself when I started contributing to Authelia. 

> Authelia is an open-source authentication and authorization server and portal fulfilling the identity and access management (IAM) role of information security in providing multi-factor authentication and single sign-on (SSO) for your applications via a web portal.

I initially deployed Authelia to protect applications that were exposed to the internet. As I used it more and more, and after reading (actually listening to) the book *Clean Code* by Robert C. Martin, who recommends all developers find an open source project to contribute to, I decided to try and contribute to Authelia.

The first real feature I contributed was [an improvement](https://github.com/authelia/authelia/commit/071be3c63281e17b568a592e181f2c993bdfea3e) to the way we display dates for 2fa credentials.

Since then, I have become one of the maintainers for Authelia and get to work with an extremely passionate and capable team of maintainers.

The deeper involvement with Authelia would eventually drive my next major infrastructure decision.

### Proxmox Clustering

With the single Rocky virtual machine working, and all its existing drives (around 6 at this point) passed through, the Proxmox system had minimal storage, and memory to spare for other virtual machines or workloads, so around this time I also increased the amount of memory in the main system (I think it was increased to 64, then later 128GB).

Also around this time, I got a Minis forum UX690X mini computer (AMD Ryzen 9 6900HX, 64GB RAM) as a second hypervisor. I had some issues with it when I tried to create a cluster and join the two Proxmox hosts.

As it turns out, proxmox nodes trying to join a cluster cannot have any running virtual machines or lxcs. If they do, the nodes will fail to join. Eventually I figured that out, saved all the VM configs, backed up all the disks, and reinstalled proxmox on my existing node. Once I did that, I was able to join the two nodes and the qDevice (a raspberry pi for quorum).

The process of backing up virtual machine configs is easy, a single folder with config files copied to a new place. 

Backing up the actual virtual machine disks was more difficult. I don't recall the method I used to backup the vms, but I did export them to a raw file and store them in the same place as the virtual machine configs (on a network share).

I did experience some data loss after the restore - my Windows VM and Kali VM didn't come back cleanly, likely due to how those systems handle disk configurations. Fortunately, these were disposable VMs I could easily rebuild, and the critical services (the Rocky Linux VM with all my containers) came through unscathed.

### Architecture at this point

- Two Proxmox hosts now called Moor (original), and Avalon (mini pc) in a cluster with a qDevice
- A number of virtual machines (~4-8)
- No LXC (Linux containers)
- [Opnsense] on a cheap purpose-built (but prebuilt) mini-pc
  - Running the HAProxy plugin to allow me to route domains (TCP passthrough) to different virtual machines based on where applications are running (using SNI inspection).
  - Running my DHCP server (ISC).
  - Running my DNS server (unbound dns plugin) for split dns (access local machines directly instead of going out to the internet and through cloudflare)

Around this time I also started naming the virtual machines and hosts that are pets [(not cattle)](https://devops.stackexchange.com/questions/653/what-is-the-definition-of-cattle-not-pets).

### Kubernetes Orchestration

Part of my continuing work on Authelia, I got more interested in Kubernetes/container orchestration environments (as Authelia is meant to be stateless). I also got tired of having to complete a stop the world event to update Authelia's configuration to add an OIDC client or update ACL rules.

The natural decision?
Deploy a Kubernetes cluster in my homelab to run Authelia in a way to allow rolling restarts and high availability.

### Kubernetes the Hard Way

Now I can't say that I poured over the Kubernetes (K8s) documentation to try and learn everything before I deployed the cluster, I didn't do that at all. I did what most people in the self-hosted community do, and found a guide.

I deployed a 3 node/3 master cluster using K3s (a lightweight Kubernetes) across my 2-node proxmox cluster. Making use of VM templates made from an ubuntu server os.

Because I didn't read the k8s documentation, I didn't understand the separation between the api server/etcd instances and the fact that Kubernetes master nodes are only needed for the control plane (to control scheduling of workloads and management of the cluster). After hitting the point where each k8s master has 8-10 cpu cores, and 24-32G of ram, I figured it was time to expand my cluster.

Looking back, this was an expensive (mostly in time) lesson in RTFM (Read the Fucking Manual). Having all workloads on a small number of nodes means that if a single node goes down, all the workloads on the missing node have to transfer to another node. When you only have two available nodes, it's really easy to run out of resources on the remaining nodes. When the remaining nodes are masters, you start running into issues scheduling pods. 

I also learned that the masters are not required to run workloads and really should only be used to run the k8s api server and etcd server.

### Automation

Manually configuring virtual machines is tedious, repetitive, and error-prone.

Rather than provisioning an unknown number of workers manually (who knows if I want to add more later), I decided I should automate it. I did some research and I came across [Flatcar] container linux, an immutable, container oriented operating system. Flatcar uses butane to automatically provision and configure itself on the first startup.

At this point, I created a butane config (compiles to ignition) to automate the initialization of k8s worker nodes. I also created a bash script that prompts for important information about the machine (hostname, static ip, dns, gateway, k3s bootstrap token), downloads the current butane config from my git server, uses placeholders to insert the important information into the config, compiles the butane config to an ignition file, creates the Proxmox virtual machine config (cpu, memory, disks, networking), and loads the config as a cloud init drive.

Then all I need to do is start the virtual machine and once it finishes initializing I have a full worker node with an external longhorn disk.

With the automation, I also started using [Flatcar] container linux for my k8s nodes (only workers for the moment). Flatcar takes immutable operating systems, and cattle not pets methodology to the next level. Flatcar vms have no package manager, if you need a package you have to install it manually or re-provision the node. Updates are handled automatically, the update is written to a background partition, then the system restarts into the background partition, with downtime only being as long as it takes to restart.

After I was able to automatically provision and initialize new k8s nodes, I started moving workloads off of the k8s masters. Today, all three k8s masters are cordoned and only run scheduling and specific workloads ([ArgoCD] controller, [Kyverno] controller, and [Longhorn] controller + replicas)

In this process, I actually added an extremely lightweight proxmox node (called Sparrow, with 4 cores, and 24G of memory) to the proxmox cluster to replace the qDevice. The main reason for this is to finally get the full benefits of running a 3-master k3s cluster (able to maintain quorum after losing a single node).

Moving the workloads off the master nodes necessarily required adding more worker nodes. In the end, I ended up with 6 worker nodes (2 on Avalon, 3 on Moor, and 1 on Sparrow), with each node also having a master node.

After moving the workloads off the master nodes, I was finally able to cordon them (prevent normal workloads from being scheduled) and downsize the resources they have (down to 2 cores, 8G memory each).

### Current Homelab Stats
- **Proxmox Hosts:** 3 (Moor, Avalon, Sparrow)
- **K8s Masters:** 3 (2 cores, 8GB RAM each)
- **K8s Workers:** 6 nodes
- **Total Pods:** ~300+
- **Storage:** Longhorn distributed across all nodes.
- **Uptime:** (as of today) Avalon: 115 Days, Moor: 5 Days, Sparrow: 29 Days

## Lessons Learned

Looking back over my journey, here are some of my key takeaways:

- **Read the documentation:** Not understanding Kubernetes architecture cost me weeks of work. Over provisioning the master nodes and the resulting extra work to migrate away from them are a reminder what happens when you jump into something unprepared.
- **Plan Big Migrations Carefully:** The issues with creating a Proxmox cluster taught me the importance of understanding things up front, some extra research on Proxmox clustering would've saved me a lot of headache and seat-of-my-pants troubleshooting.
- **Automate early and often:** While this didn't come from a specific issue, its still a good idea. When you are managing infrastructure, it's much better both in time spent and brainpower to automate the repetitive tasks up front. This ensures you are able to repeat what you did with the certainty that you didn't miss anything. I must have reprovisioned that first worker node 5-10 times before I got the ignition script the way I wanted and all the kinks worked out.
- **Document everything:** This ties into the previous point. By documenting everything, you can look back at what you did to solve a specific issue. Especially in a homelab, where you are learning and solving problems for the first time. It's really hard to remember absolutely every command and step you took to do something. Looking back, I really wish I had documented my earlier days more.
- **Expect to fail:** Every mistake I have made has taught me something. Whether that's you can't create a proxmox cluster with nodes already running vms, or that infrastructure dependency chains are bad. The point of a homelab is to make these mistakes in a semi-safe environment that doesn't affect paying customers.

One of the really rewarding things is to know that I am using tools that are used in the enterprise space. ArgoCD, K3s/K8s, and Longhorn bring enterprise-grade software into my homelab.

## What's Next

Currently, I am working on adding mTLS to my lab. A lot of the infrastructure I have mainly uses normal server TLS and relies on the client to validate they are connecting to the right server. I am working on setting up mTLS on all my machines and clients.

In the next post, I will likely cover what is actually running in my lab today (monitoring, databases, authentication, and the applications I am using). I will likely also discuss monitoring, backups, recovery, and any maintenance.


[apache2]: https://httpd.apache.org/
[Traefik]: https://traefik.io
[K3s]: https://k3s.io/
[Loki]: https://grafana.com/docs/loki/latest/
[Grafana]: https://grafana.com/docs/grafana/latest/
[Tempo]: https://grafana.com/docs/tempo/latest/
[Mimir]: https://grafana.com/docs/mimir/latest/
[Pyroscope]: https://grafana.com/docs/pyroscope/latest/
[Cloud-native PostgreSQL]: https://cloudnative-pg.io/
[Scale sets]: https://docs.github.com/en/actions/tutorials/use-actions-runner-controller/deploy-runner-scale-sets
[Longhorn]: https://longhorn.io/
[ArgoCD]: https://argo-cd.readthedocs.io/en/stable/
[Kyverno]: https://kyverno.io/
[Opnsense]: https://opnsense.org/
[Authelia]: https://authelia.com/
[Cert-manager]: https://cert-manager.io/
[Flatcar]: https://www.flatcar.org/