$module('async', function () {
  var $static = this;
  $static.DoSomethingAsync = $t.workerwrap('8a260667', function (a) {
    return a;
  });
  $static.TEST = $t.markpromising(function () {
    var $result;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            $promise.translate($g.async.DoSomethingAsync($t.fastbox(3, $g.________testlib.basictypes.Integer))).then(function ($result0) {
              $result = $t.fastbox($result0.$wrapped == 3, $g.________testlib.basictypes.Boolean);
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
