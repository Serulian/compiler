$module('init', function () {
  var $static = this;
  $static.TEST = function () {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $g.____testlib.basictypes.Integer.$equals($g.init.sc.value, $t.box(2, $g.____testlib.basictypes.Integer)).then(function ($result1) {
              return $promise.resolve($t.unbox($result1)).then(function ($result0) {
                return ($promise.shortcircuit($result0, true) || $g.____testlib.basictypes.Integer.$equals($g.init.sc2.value, $t.box(4, $g.____testlib.basictypes.Integer))).then(function ($result2) {
                  $result = $t.box($result0 && $t.unbox($result2), $g.____testlib.basictypes.Boolean);
                  $current = 1;
                  $continue($resolve, $reject);
                  return;
                });
              });
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
  };
  this.$init(function () {
    return $g.other.SomeClass.NewThing($t.box(1, $g.____testlib.basictypes.Integer)).then(function ($result0) {
      $static.sc = $result0;
    });
  }, '194711ba', ['cc19450d']);
  this.$init(function () {
    return $g.other.SomeClass.NewThing($t.box(2, $g.____testlib.basictypes.Integer)).then(function ($result0) {
      $static.sc2 = $result0;
    });
  }, '2e80db22', ['cc19450d']);
});