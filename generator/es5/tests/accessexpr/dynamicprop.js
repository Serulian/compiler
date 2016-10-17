$module('dynamicprop', function () {
  var $static = this;
  this.$class('SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      instance.value = $t.box(42, $g.____testlib.basictypes.Integer);
      return $promise.resolve(instance);
    };
    $instance.SomeProp = $t.property(function () {
      var $this = this;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        $resolve($this.value);
        return;
      };
      return $promise.new($continue);
    }, function (val) {
      var $this = this;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        $this.value = val;
        $resolve();
        return;
      };
      return $promise.new($continue);
    });
    this.$typesig = function () {
      return $t.createtypesig(['SomeProp', 3, $g.____testlib.basictypes.Integer.$typeref()], ['new', 1, $g.____testlib.basictypes.Function($g.dynamicprop.SomeClass).$typeref()]);
    };
  });

  $static.TEST = function () {
    var $result;
    var sc;
    var sca;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $g.dynamicprop.SomeClass.new().then(function ($result0) {
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
            sc = $result;
            sc.SomeProp($t.box(123, $g.____testlib.basictypes.Integer)).then(function ($result0) {
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
            sca = sc;
            $t.dynamicaccess(sca, 'SomeProp').then(function ($result2) {
              return $g.____testlib.basictypes.Integer.$equals($t.cast($result2, $g.____testlib.basictypes.Integer, false), $t.box(123, $g.____testlib.basictypes.Integer)).then(function ($result1) {
                return $promise.resolve($t.unbox($result1)).then(function ($result0) {
                  return ($promise.shortcircuit($result0, true) || sc.SomeProp()).then(function ($result4) {
                    return ($promise.shortcircuit($result0, true) || $g.____testlib.basictypes.Integer.$equals($result4, $t.box(123, $g.____testlib.basictypes.Integer))).then(function ($result3) {
                      $result = $t.box($result0 && $t.unbox($result3), $g.____testlib.basictypes.Boolean);
                      $current = 3;
                      $continue($resolve, $reject);
                      return;
                    });
                  });
                });
              });
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 3:
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
