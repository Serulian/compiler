$module('withexit', function () {
  var $static = this;
  this.$class('1ac0891b', 'SomeReleasable', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.Release = function () {
      var $this = this;
      $g.withexit.someBool = $t.fastbox(true, $g.________testlib.basictypes.Boolean);
      return;
    };
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

  $static.TEST = function () {
    var $temp0;
    var $current = 0;
    var $resources = $t.resourcehandler();
    syncloop: while (true) {
      switch ($current) {
        case 0:
          $t.fastbox(123, $g.________testlib.basictypes.Integer);
          $current = 1;
          continue syncloop;

        case 1:
          $temp0 = $g.withexit.SomeReleasable.new();
          $resources.pushr($temp0, '$temp0');
          $t.fastbox(456, $g.________testlib.basictypes.Integer);
          if (false) {
            $current = 2;
            continue syncloop;
          } else {
            $current = 4;
            continue syncloop;
          }
          break;

        case 2:
          $current = 3;
          continue syncloop;

        case 3:
          $t.fastbox(789, $g.________testlib.basictypes.Integer);
          var $pat = $g.withexit.someBool;
          $resources.popall();
          return $pat;

        case 4:
          $t.fastbox(12, $g.________testlib.basictypes.Integer);
          $resources.popr('$temp0');
          var $pat = $g.withexit.someBool;
          $resources.popall();
          return $pat;

        default:
          return;
      }
    }
  };
  this.$init(function () {
    return $promise.new(function (resolve) {
      $static.someBool = $t.fastbox(false, $g.________testlib.basictypes.Boolean);
      resolve();
    });
  }, '0b58b8ac', []);
});
