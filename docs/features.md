Flagship: Barrelman Features
====
[//]: # (Checked = &#9745;    Unchecked = &#9744;)    
| Feature | Description | Implemented |
|:---------:|:------:|:-------:|
Multiple Chart deployment | Any number of Helm charts may be defined in a Barrelment manifest to be deployed in an atomic manner, potentially across namespaces. | &#9745; |
Automatic Chart Building | Barrelman can routinely build Helm charts assembled from constituent sources at runtime as defined in a Barrelman manifest. |   &#9745; |
Atomic apply command| Barrelman will analyze current state of the cluster, compute the necessary actions, deploy the computed plan, and react to changes in Kubernetes on the fly in order to achieve the target deployment state. | &#9745; |
Delete command | This command is the counterpart to the "Apply" command and will delete all releases defined in the Barrelman manifest. | &#9745; |
Multiple chart source protocols | Barrelman supports multiple chart source locations such as Git and local directories. These can be used at the same time seamlessly to assemble Charts. | &#9745;
Diff option | The differences between the current cluster state and the proposed action can be displayed an an easy to read format. | &#9745;
Rollback command | A state change can be rolled back using the rollback information stored in the kubernetes cluster, the releases specified in the manifest will be rolled back in a constistent manner to treat the application stack as a unified state. | &#9744;
ConfigMap lifecycle management | Barrelman can "Apply" and Delete Kubernetes ConfigMaps as part of the application stack lifecycle. | &#9744;
Pre/Post hooks | Jobs can be run to analyze and block execution until coded conditions are met. This is useful to wait until services have completed initialization before Barrelman can proceed to the following steps, and to potentially run conditional initialization jobs such as populating a database. | &#9744;
Backup/Restore | Barrelman can snapshot and export to file a running state of a cluster and restore that state at a later time, potentially allowing for modification by an operator before the restore. | &#9744;
HTTP/s Manifest loader | While it is common to operate Barrelman with a local manifest file to deploy and update an application stack, Barrelman can also retrieve a manifest from a secure web server and run the retrieved manifest as if it were local. | &#9744;
---
## Git Source Handler
Git source endpoints are supported with the following features: 
- GitHub private repo 
- Multiple repository
- Checkout specific branch
- Checkout specific tag
- Usable with the automatic chart build feature
- Automatic synchronization (git pull) during "apply"<br/>

## Directory Source Handler
Local directory can be used as a chart source repository with the following features:
- Usable with the automatic chart build feature
- Efficient Modify, Run/Test workflow
- Use with your favorite git workflow

## File source handler
A local Helm Chart archive can be used directly from a Barrelman manifest.
