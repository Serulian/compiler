$module('interfacecastfail', function () {
  var $static = this;
  this.$class('90a7ad6b', 'SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.SomeValue = $t.property(function () {
      var $this = this;
      return $t.fastbox(2, $g.________testlib.basictypes.Integer);
    });
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "SomeValue|3|2e508ae6": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  this.$interface('470545ee', 'SomeInterface', false, '', function () {
    var $static = this;
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "SomeValue|3|aa28dc2d": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.TEST = function () {
    var sc;
    sc = $g.interfacecastfail.SomeClass.new();
    $t.cast(sc, $g.interfacecastfail.SomeInterface, false);
    return;
  };
});
