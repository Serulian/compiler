$module('nullbox', function () {
  var $static = this;
  this.$struct('45b3c939', 'SomeStruct', false, '', function () {
    var $static = this;
    var $instance = this.prototype;
    $static.new = function () {
      var instance = new $static();
      instance[BOXED_DATA_PROPERTY] = {
      };
      instance.$markruntimecreated();
      return instance;
    };
    $static.$fields = [];
    $t.defineStructField($static, 'Value', 'Value', function () {
      return $g.________testlib.basictypes.String;
    }, function () {
      return $g.________testlib.basictypes.String;
    }, true);
    this.$typesig = function () {
      if (this.$cachedtypesig) {
        return this.$cachedtypesig;
      }
      var computed = {
        "Parse|1|2549c819<45b3c939>": true,
        "equals|4|2549c819<f361570c>": true,
        "Stringify|2|2549c819<ec87fc3f>": true,
        "Mapping|2|2549c819<95776681<any>>": true,
        "Clone|2|2549c819<45b3c939>": true,
        "String|2|2549c819<ec87fc3f>": true,
      };
      return this.$cachedtypesig = computed;
    };
  });

  $static.TEST = function () {
    var $temp0;
    var s;
    s = ($temp0 = $g.nullbox.SomeStruct.new(), $temp0.Value = null, $temp0);
    return $t.fastbox((s.Value == null) && (s.Mapping().$index($t.fastbox('Value', $g.________testlib.basictypes.String)) == null), $g.________testlib.basictypes.Boolean);
  };
});
