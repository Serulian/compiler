$module('constructable', function () {
  var $static = this;
  this.$class('ceb86b8e', 'SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $static.Get = function () {
      return $g.constructable.SomeClass.new();
    };
    $instance.SomeBool = $t.property(function () {
      var $this = this;
      return $t.fastbox(true, $g.____testlib.basictypes.Boolean);
    });
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Get|1|89b8f38e<ceb86b8e>": true,
        "SomeBool|3|f7f23c49": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  this.$interface('ba50ace2', 'SomeInterface', false, '', function () {
    var $static = this;
    $static.Get = function () {
      return $g.constructable.SomeClass.new();
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Get|1|89b8f38e<ba50ace2>": true,
        "SomeBool|3|f7f23c49": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.DoSomething = function (T) {
    var $f = function () {
      return T.Get();
    };
    return $f;
  };
  $static.TEST = function () {
    return $g.constructable.DoSomething($g.constructable.SomeInterface)().SomeBool();
  };
});
