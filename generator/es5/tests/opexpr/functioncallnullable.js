$module('functioncallnullable', function () {
  var $static = this;
  this.$class('fa2a0786', 'SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.SomeMethod = function () {
      var $this = this;
      return $t.fastbox(true, $g.____testlib.basictypes.Boolean);
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "SomeMethod|2|89b8f38e<f7f23c49>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  this.$class('97cf7b5c', 'AnotherClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.AnotherMethod = function () {
      var $this = this;
      return $t.fastbox(false, $g.____testlib.basictypes.Boolean);
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "AnotherMethod|2|89b8f38e<f7f23c49>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.TEST = function () {
    var ac;
    var sc;
    sc = $g.functioncallnullable.SomeClass.new();
    ac = null;
    return $t.fastbox($t.syncnullcompare($t.nullableinvoke(sc, 'SomeMethod', false, []), function () {
      return $t.fastbox(false, $g.____testlib.basictypes.Boolean);
    }).$wrapped && $t.syncnullcompare($t.nullableinvoke(ac, 'AnotherMethod', false, []), function () {
      return $t.fastbox(true, $g.____testlib.basictypes.Boolean);
    }).$wrapped, $g.____testlib.basictypes.Boolean);
  };
});
