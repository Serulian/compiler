$module('jsondefault', function () {
  var $static = this;
  this.$struct('d610e0b2', 'SomeStruct', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      instance[BOXED_DATA_PROPERTY] = {
      };
      instance.$markruntimecreated();
      return $static.$initDefaults(instance, true);
    };
    $static.$initDefaults = function (instance, isRuntimeCreated) {
      var boxed = instance[BOXED_DATA_PROPERTY];
      if (isRuntimeCreated || (boxed['SomeField'] === undefined)) {
        instance.SomeField = $t.fastbox(42, $g.____testlib.basictypes.Integer);
      }
      return instance;
    };
    $static.$fields = [];
    $t.defineStructField($static, 'SomeField', 'SomeField', function () {
      return $g.____testlib.basictypes.Integer;
    }, function () {
      return $g.____testlib.basictypes.Integer;
    }, false);
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Parse|1|89b8f38e<d610e0b2>": true,
        "equals|4|89b8f38e<f7f23c49>": true,
        "Stringify|2|89b8f38e<549fbddd>": true,
        "Mapping|2|89b8f38e<ad6de9ce<any>>": true,
        "Clone|2|89b8f38e<d610e0b2>": true,
        "String|2|89b8f38e<549fbddd>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.TEST = $t.markpromising(function () {
    var $result;
    var jsonString;
    var parsed;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      while (true) {
        switch ($current) {
          case 0:
            jsonString = $t.fastbox('{}', $g.____testlib.basictypes.String);
            $promise.maybe($g.jsondefault.SomeStruct.Parse($g.____testlib.basictypes.JSON)(jsonString)).then(function ($result0) {
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
            parsed = $result;
            $resolve($t.fastbox(parsed.SomeField.$wrapped == 42, $g.____testlib.basictypes.Boolean));
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
