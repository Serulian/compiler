$module('basic', function () {
  var $static = this;
  this.$class('da5f206e', 'SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      return instance;
    };
    $instance.DoSomething = function () {
      var $this = this;
      return $t.fastbox(true, $g.________testlib.basictypes.Boolean);
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "DoSomething|2|fd8bc7c9<54ff3ddf>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  this.$type('42b8d372', 'MyType', false, '', function () {
    var $instance = this.prototype;
    var $static = this;
    this.$box = function ($wrapped) {
      var instance = new this();
      instance[BOXED_DATA_PROPERTY] = $wrapped;
      return instance;
    };
    this.$roottype = function () {
      return $g.basic.SomeClass;
    };
    $instance.AnotherThing = function () {
      var $this = this;
      return $this.$wrapped.DoSomething();
    };
    $instance.SomeProp = $t.property(function () {
      var $this = this;
      return $t.fastbox(true, $g.________testlib.basictypes.Boolean);
    });
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "AnotherThing|2|fd8bc7c9<54ff3ddf>": true,
        "SomeProp|3|54ff3ddf": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.TEST = function () {
    var m;
    var sc;
    sc = $g.basic.SomeClass.new();
    m = $t.fastbox(sc, $g.basic.MyType);
    return $t.fastbox(m.SomeProp().$wrapped && m.AnotherThing().$wrapped, $g.________testlib.basictypes.Boolean);
  };
});
