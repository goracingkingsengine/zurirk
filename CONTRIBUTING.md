Contributions to zurichess are welcomed. These are some guidelines to follow:

* Keep the patch as simple as possible. Get familiar with this article
http://tirania.org/blog/archive/2010/Dec-31.html.
* go fmt and go test your code. Add tests if possible.
* The code will be reviewed. Complicated patches will be rejected, but
feedback will be given. Changes that break encapsulation boundaries (e.g.
adding evaluation elements to the board logic) will be rejected.
* Do not manually tune evaluation parameters. It is a waste of time, use
txt instead. Search parameters can be tuned manually since there are only
a few of them. Generally prefer to use automatic tools for tuning.
* Any patch will be rigurously tested at two different time controls. To
reduce cluster testing time please include in the pull request the results
of a match of at least 5000 games at 40/1+0.025.

Things that can be improved:

* Evaluation is very basic and needs more features.
* Move ordering only considers hash, killers and MVV/LVA.
* LMR, NMP and FP conditions are not very well tuned.
* SEE prunings in QS misses some tactics such as discovered attacks.
* No SMP, zurichess uses a single thread.
