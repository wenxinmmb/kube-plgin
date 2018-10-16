# Create, Update & Delete Deployment

This example program demonstrates the fundamental operations for managing on
[Deployment][1] resources, such as `Create`, `List`, `Update` and `Delete`.

```
Running this command will execute the following operations on your cluster:

1. **Create Deployment:** This will create a 2 replica Deployment. Verify with
   `kubectl get pods`.
2. **Update Deployment:** This will update the Deployment resource created in
   previous step by setting the replica count to 1 and changing the container
   image to `nginx:1.13`. You are encouraged to inspect the retry loop that
   handles conflicts. Verify the new replica count and container image with
   `kubectl describe deployment demo`.
3. **Rollback Deployment:** This will rollback the Deployment to the last
   revision. In this case, it's the revision that was created in Step 1.
   Use `kubectl describe` to verify the container image is now `nginx:1.12`.
   Also note that the Deployment's replica count is still 1; this is because a
   Deployment revision is created if and only if the Deployment's pod template
   (`.spec.template`) is changed.
4. **List Deployments:** This will retrieve Deployments in the `default`
   namespace and print their names and replica counts.
5. **Delete Deployment:** This will delete the Deployment object and its
   dependent ReplicaSet resource. Verify with `kubectl get deployments`.

Each step is separated by an interactive prompt. You must hit the
<kbd>Return</kbd> key to proceed to the next step. You can use these prompts as
a break to take time to  run `kubectl` and inspect the result of the operations
executed.

You should see an output like the following:

```
Creating deployment...
Created deployment "demo-deployment".
-> Press Return key to continue.

Updating deployment...
Updated deployment...
-> Press Return key to continue.

Rolling back deployment...
Rolled back deployment...
-> Press Return key to continue.

Listing deployments in namespace "default":
 * demo-deployment (1 replicas)
-> Press Return key to continue.

Deleting deployment...
Deleted deployment.
```


