$module('isnull', function () {
  var $static = this;
  $static.DoSomething = function (a) {
    var $state = $t.sm(function ($callback) {
      while (true) {
        switch ($state.current) {
          case 0:
            $state.resolve($t.nominalwrap(a == null, $g.____testlib.basictypes.Boolean));
            return;

          default:
            $state.current = -1;
            return;
        }
      }
    });
    return $promise.build($state);
  };
});
