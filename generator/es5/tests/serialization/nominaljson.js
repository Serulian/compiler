$module('nominaljson', function () {
  var $static = this;
  this.$type('4a09c49b', 'SomeNominal', false, '', function () {
    var $instance = this.prototype;
    var $static = this;
    this.$box = function ($wrapped) {
      var instance = new this();
      instance[BOXED_DATA_PROPERTY] = $wrapped;
      return instance;
    };
    this.$roottype = function () {
      return $g.nominaljson.AnotherStruct;
    };
    $instance.GetValue = function () {
      var $this = this;
      return $t.box($this, $g.nominaljson.AnotherStruct).AnotherBool;
    };
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "GetValue|2|cf412abd<aa28dc2d>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  this.$struct('c267cdf9', 'AnotherStruct', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function (AnotherBool) {
      var instance = new $static();
      instance[BOXED_DATA_PROPERTY] = {
        AnotherBool: AnotherBool,
      };
      instance.$markruntimecreated();
      return instance;
    };
    $static.$fields = [];
    $t.defineStructField($static, 'AnotherBool', 'AnotherBool', function () {
      return $g.________testlib.basictypes.Boolean;
    }, function () {
      return $g.________testlib.basictypes.Boolean;
    }, false);
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Parse|1|cf412abd<c267cdf9>": true,
        "equals|4|cf412abd<aa28dc2d>": true,
        "Stringify|2|cf412abd<cb470bcc>": true,
        "Mapping|2|cf412abd<899aec48<any>>": true,
        "Clone|2|cf412abd<c267cdf9>": true,
        "String|2|cf412abd<cb470bcc>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  this.$struct('eda8a36d', 'SomeStruct', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function (Nested) {
      var instance = new $static();
      instance[BOXED_DATA_PROPERTY] = {
        Nested: Nested,
      };
      instance.$markruntimecreated();
      return instance;
    };
    $static.$fields = [];
    $t.defineStructField($static, 'Nested', 'Nested', function () {
      return $g.nominaljson.SomeNominal;
    }, function () {
      return $g.nominaljson.SomeNominal;
    }, false);
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Parse|1|cf412abd<eda8a36d>": true,
        "equals|4|cf412abd<aa28dc2d>": true,
        "Stringify|2|cf412abd<cb470bcc>": true,
        "Mapping|2|cf412abd<899aec48<any>>": true,
        "Clone|2|cf412abd<eda8a36d>": true,
        "String|2|cf412abd<cb470bcc>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.TEST = $t.markpromising(function () {
    var $result;
    var correct;
    var jsonString;
    var parsed;
    var s;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            s = $g.nominaljson.SomeStruct.new($t.box($g.nominaljson.AnotherStruct.new($t.fastbox(true, $g.________testlib.basictypes.Boolean)), $g.nominaljson.SomeNominal));
            jsonString = $t.fastbox('{"Nested":{"AnotherBool":true}}', $g.________testlib.basictypes.String);
            $promise.maybe(s.Stringify($g.________testlib.basictypes.JSON)()).then(function ($result0) {
              $result = $g.________testlib.basictypes.String.$equals($result0, jsonString);
              $current = 1;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 1:
            correct = $result;
            $promise.maybe($g.nominaljson.SomeStruct.Parse($g.________testlib.basictypes.JSON)(jsonString)).then(function ($result0) {
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
            parsed = $result;
            $resolve($t.fastbox(correct.$wrapped && parsed.Nested.GetValue().$wrapped, $g.________testlib.basictypes.Boolean));
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
