$module('withasync', function () {
  var $static = this;
  this.$class('74cb3efd', 'SomeReleasable', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.Release = $t.markpromising(function () {
      var $this = this;
      var $result;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        localasyncloop: while (true) {
          switch ($current) {
            case 0:
              $promise.translate($g.withasync.DoSomethingAsync()).then(function ($result0) {
                $result = $g.withasync.someBool = $result0;
                $current = 1;
                $continue($resolve, $reject);
                return;
              }).catch(function (err) {
                $reject(err);
                return;
              });
              return;

            case 1:
              $resolve();
              return;

            default:
              $resolve();
              return;
          }
        }
      };
      return $promise.new($continue);
    });
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Release|2|cf412abd<void>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.DoSomethingAsync = $t.workerwrap('70fe88e1', function () {
    return $t.fastbox(true, $g.________testlib.basictypes.Boolean);
  });
  $static.TEST = $t.markpromising(function () {
    var $result;
    var $temp0;
    var $current = 0;
    var $resources = $t.resourcehandler();
    var $continue = function ($resolve, $reject) {
      $resolve = $resources.bind($resolve, true);
      $reject = $resources.bind($reject, true);
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            $t.fastbox(123, $g.________testlib.basictypes.Integer);
            $temp0 = $g.withasync.SomeReleasable.new();
            $resources.pushr($temp0, '$temp0');
            $t.fastbox(456, $g.________testlib.basictypes.Integer);
            $resources.popr('$temp0').then(function ($result0) {
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
            $t.fastbox(789, $g.________testlib.basictypes.Integer);
            $resolve($g.withasync.someBool);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  });
  this.$init(function () {
    return $promise.new(function (resolve) {
      $static.someBool = $t.fastbox(false, $g.________testlib.basictypes.Boolean);
      resolve();
    });
  }, '4a6074dd', []);
});
