#### Seamless upgrade of the operator and the operand
-

---
#### Compare the current version of the CR instance with the container image.

```
	if bestieImageDifferent {
		log.Info("Upgrade Operand")
		dp.Spec.Template.Spec.Containers[0].Image = getBestieContainerImage(bestie)
		err = r.Client.Update(ctx, dp)
		if err != nil {
			log.Error(err, "Need to update, but failed to update bestie image")
			return err
		}
	}
```
---
#### Slide 3
