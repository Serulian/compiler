$module('expectrejection', function () {
  var $static = this;
  this.$class('aff287b8', 'SimpleError', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return $promise.resolve(instance);
    };
    $instance.Message = $t.property(function () {
      var $this = this;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        $resolve($t.fastbox('yo!', $g.____testlib.basictypes.String));
        return;
      };
      return $promise.new($continue);
    });
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Message|3|538656f2": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.DoSomething = function () {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $g.expectrejection.SimpleError.new().then(function ($result0) {
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
            $reject($result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  };
  $static.TEST = function () {
    var $result;
    var a;
    var b;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $g.expectrejection.DoSomething().then(function ($result0) {
              a = $result0;
              b = null;
              $current = 1;
              $continue($resolve, $reject);
              return;
            }).catch(function ($rejected) {
              b = $rejected;
              a = null;
              $current = 1;
              $continue($resolve, $reject);
              return;
            });
            return;

          case 1:
            $promise.resolve(a == null).then(function ($result0) {
              return ($promise.shortcircuit($result0, true) || $t.assertnotnull(b).Message()).then(function ($result2) {
                return ($promise.shortcircuit($result0, true) || $g.____testlib.basictypes.String.$equals($result2, $t.fastbox('yo!', $g.____testlib.basictypes.String))).then(function ($result1) {
                  $result = $t.fastbox($result0 && $result1.$wrapped, $g.____testlib.basictypes.Boolean);
                  $current = 2;
                  $continue($resolve, $reject);
                  return;
                });
              });
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 2:
            $resolve($result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  };
});
