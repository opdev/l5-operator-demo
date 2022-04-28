#### You deserve a good night's sleep
Can your operator:
- Read metrics?
- Auto-scale?
- Move Workloads?
- Restart internal resources?
---
#### Enabling HPA
- Set MaxReplicas to activate it.
````
func horizontalpodautoscalers(ctx context.Context, bestieDeployment appsv1.Deployment, bestie v1.Bestie, client cli.Client, r *runtime.Scheme) error {
	desired := []autoscalingv1.HorizontalPodAutoscaler{}

	if bestie.Spec.MaxReplicas != nil {
		log.Info("MaxReplicas is set, enabling HPA")
		desired = append(desired, hpa.AutoScaler(ctrllog.Log, bestieDeployment, bestie))
	}

	if err := applyHorizontalPodAutoscalers(ctx, bestie, client, r, desired); err != nil {
		log.Error(err, "failed to reconcile the expected horizontal pod autoscalers")
		return err
	}
	return nil
}
````


---
#### Fake load Demo

kubectl run -i --tty load-generator --rm --image=busybox:1.28 --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://bestie-service; done"
````
kubectl run -i --tty load-generator --rm --image=busybox:1.28 --restart=Never -- /bin/sh -c "while sleep 0.01; do wget -q -O- http://bestie-service; done"
````
---
#### Thank you!