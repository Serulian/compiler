$module('property', function () {
  var $static = this;
  this.$class('SomeClass', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      instance.SomeBool = $t.box(false, $g.____testlib.basictypes.Boolean);
      return $promise.resolve(instance);
    };
    $instance.set$SomeProp = function (val) {
      var $this = this;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        $this.SomeBool = val;
        $resolve();
        return;
      };
      return $promise.new($continue);
    };
    $instance.SomeProp = $t.property(function () {
      var $this = this;
      var $current = 0;
      var $continue = function ($resolve, $reject) {
        $resolve($this.SomeBool);
        return;
      };
      return $promise.new($continue);
    });
    this.$typesig = function () {
      return $t.createtypesig(['SomeProp', 3, $g.____testlib.basictypes.Boolean.$typeref()], ['new', 1, $g.____testlib.basictypes.Function($g.property.SomeClass).$typeref()]);
    };
  });

  $static.AnotherFunction = function (sc) {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            sc.SomeProp().then(function ($result0) {
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
            sc.set$SomeProp($t.box(true, $g.____testlib.basictypes.Boolean)).then(function ($result0) {
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
            sc.SomeProp().then(function ($result0) {
              $result = $result0;
              $current = 3;
              $continue($resolve, $reject);
              return;
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
  $static.TEST = function () {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $g.property.SomeClass.new().then(function ($result1) {
              return $g.property.AnotherFunction($result1).then(function ($result0) {
                $result = $result0;
                $current = 1;
                $continue($resolve, $reject);
                return;
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
});
