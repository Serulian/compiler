$module('clone', function () {
  var $static = this;
  this.$struct('SomeStruct', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function (SomeField, AnotherField) {
      var instance = new $static();
      instance.$unboxed = false;
      instance[BOXED_DATA_PROPERTY] = {
        SomeField: SomeField,
        AnotherField: AnotherField,
      };
      return $promise.resolve(instance);
    };
    $static.$fields = [];
    $t.defineStructField($static, 'SomeField', 'SomeField', function () {
      return $g.____testlib.basictypes.Integer;
    }, true, function () {
      return $g.____testlib.basictypes.Integer;
    }, false);
    $t.defineStructField($static, 'AnotherField', 'AnotherField', function () {
      return $g.____testlib.basictypes.Boolean;
    }, true, function () {
      return $g.____testlib.basictypes.Boolean;
    }, false);
    this.$typesig = function () {
      return $t.createtypesig(['new', 1, $g.____testlib.basictypes.Function($g.clone.SomeStruct).$typeref()], ['Parse', 1, $g.____testlib.basictypes.Function($g.clone.SomeStruct).$typeref()], ['equals', 4, $g.____testlib.basictypes.Function($g.____testlib.basictypes.Boolean).$typeref()], ['Stringify', 2, $g.____testlib.basictypes.Function($g.____testlib.basictypes.String).$typeref()], ['Mapping', 2, $g.____testlib.basictypes.Function($g.____testlib.basictypes.Mapping($t.any)).$typeref()], ['Clone', 2, $g.____testlib.basictypes.Function($g.clone.SomeStruct).$typeref()], ['String', 2, $g.____testlib.basictypes.Function($g.____testlib.basictypes.String).$typeref()]);
    };
  });

  $static.TEST = function () {
    var first;
    var second;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            $g.clone.SomeStruct.new($t.box(42, $g.____testlib.basictypes.Integer), $t.box(false, $g.____testlib.basictypes.Boolean)).then(function ($result0) {
              $temp0 = $result0;
              $result = ($temp0, $temp0);
              $current = 1;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 1:
            first = $result;
            first.Clone().then(function ($result0) {
              $temp1 = $result0;
              $result = ($temp1, $temp1.AnotherField = $t.box(true, $g.____testlib.basictypes.Boolean), $temp1);
              $current = 2;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 2:
            second = $result;
            $g.____testlib.basictypes.Integer.$equals(second.SomeField, $t.box(42, $g.____testlib.basictypes.Integer)).then(function ($result2) {
              return $promise.resolve($t.unbox($result2)).then(function ($result1) {
                return $promise.resolve($result1 && $t.unbox(second.AnotherField)).then(function ($result0) {
                  $result = $t.box($result0 && !$t.unbox(first.AnotherField), $g.____testlib.basictypes.Boolean);
                  $current = 3;
                  $continue($resolve, $reject);
                  return;
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