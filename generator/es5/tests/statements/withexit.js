$module('withexit', function () {
  var $static = this;
  this.$class('SomeReleasable', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      var init = [];
      return $promise.all(init).then(function () {
        return instance;
      });
    };
    $instance.Release = function () {
      var $this = this;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        $g.withexit.someBool = $t.box(true, $g.____testlib.basictypes.Boolean);
        $resolve();
        return;
      };
      return $promise.new($continue);
    };
    this.$typesig = function () {
      return $t.createtypesig(['Release', 2, $g.____testlib.basictypes.Function($t.void).$typeref()], ['new', 1, $g.____testlib.basictypes.Function($g.withexit.SomeReleasable).$typeref()]);
    };
  });

  $static.TEST = function () {
    var $temp0;
    var $current = 0;
    var $resources = $t.resourcehandler();
    var $continue = function ($resolve, $reject) {
      $resolve = $resources.bind($resolve);
      $reject = $resources.bind($reject);
      while (true) {
        switch ($current) {
          case 0:
            $t.box(123, $g.____testlib.basictypes.Integer);
            $current = 1;
            continue;

          case 1:
            $g.withexit.SomeReleasable.new().then(function ($result0) {
              $result = $result0;
              $current = 2;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 2:
            $temp0 = $result;
            $resources.pushr($temp0, '$temp0');
            $t.box(456, $g.____testlib.basictypes.Integer);
            if (false) {
              $current = 3;
              continue;
            } else {
              $current = 5;
              continue;
            }
            break;

          case 3:
            $resources.popr('$temp0').then(function () {
              $current = 4;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            continue;

          case 5:
            $t.box(12, $g.____testlib.basictypes.Integer);
            $resources.popr('$temp0').then(function ($result0) {
              $result = $result0;
              $current = 6;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 6:
            $result;
            $resolve($g.withexit.someBool);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  };
  this.$init(function () {
    return $promise.resolve($t.box(false, $g.____testlib.basictypes.Boolean)).then(function (result) {
      $static.someBool = result;
    });
  });
});
