$module('await', function () {
  var $static = this;
  this.$class('15381f2b', 'SomePromise', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.Then = $t.markpromising(function (resolve) {
      var $this = this;
      var $result;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        localasyncloop: while (true) {
          switch ($current) {
            case 0:
              $promise.maybe(resolve($t.fastbox(true, $g.________testlib.basictypes.Boolean))).then(function ($result0) {
                $result = $result0;
                $current = 1;
                $continue($resolve, $reject);
                return;
              }).catch(function (err) {
                $reject(err);
                return;
              });
              return;

            case 1:
              $resolve($this);
              return;

            default:
              $resolve();
              return;
          }
        }
      };
      return $promise.new($continue);
    });
    $instance.Catch = function (rejection) {
      var $this = this;
      return $this;
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Then|2|6caba86c<12e364e4<0e92a8bc>>": true,
        "Catch|2|6caba86c<12e364e4<0e92a8bc>>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.DoSomething = $t.markpromising(function (p) {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            $promise.translate(p).then(function ($result0) {
              $result = $result0;
              $current = 1;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 1:
            $resolve($result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  });
  $static.TEST = $t.markpromising(function () {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            $promise.maybe($g.await.DoSomething($g.await.SomePromise.new())).then(function ($result0) {
              $result = $result0;
              $current = 1;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 1:
            $resolve($result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  });
});
