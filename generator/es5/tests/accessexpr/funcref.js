$module('funcref', function () {
  var $static = this;
  this.$class('41fdfa22', 'SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function (value) {
      var instance = new $static();
      instance.value = value;
      return instance;
    };
    $instance.SomeFunction = function () {
      var $this = this;
      return $this.value;
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "SomeFunction|2|cf412abd<aa28dc2d>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.AnotherFunction = $t.markpromising(function (toCall) {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            $promise.maybe(toCall()).then(function ($result0) {
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
            $resolve($result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  });
  $static.TEST = $t.markpromising(function () {
    var $result;
    var sc;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            sc = $g.funcref.SomeClass.new($t.fastbox(true, $g.________testlib.basictypes.Boolean));
            $promise.maybe($g.funcref.AnotherFunction($t.dynamicaccess(sc, 'SomeFunction', false))).then(function ($result0) {
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
            $resolve($result);
            return;

          default:
            $resolve();
            return;
        }
      }
    };
    return $promise.new($continue);
  });
});
