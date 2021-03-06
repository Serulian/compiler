$module('nested', function () {
  var $static = this;
  $static.AnotherGenerator = function () {
    var $current = 0;
    var $continue = function ($yield, $yieldin, $reject, $done) {
      while (true) {
        switch ($current) {
          case 0:
            $yield($t.fastbox(1, $g.________testlib.basictypes.Integer));
            $current = 1;
            return;

          case 1:
            if (true) {
              $current = 2;
              $continue($yield, $yieldin, $reject, $done);
              return;
            } else {
              $current = 6;
              $continue($yield, $yieldin, $reject, $done);
              return;
            }
            break;

          case 2:
            $yield($t.fastbox(2, $g.________testlib.basictypes.Integer));
            $current = 3;
            return;

          case 3:
            $current = 4;
            $continue($yield, $yieldin, $reject, $done);
            return;

          case 4:
            $yield($t.fastbox(4, $g.________testlib.basictypes.Integer));
            $current = 5;
            return;

          case 6:
            $yield($t.fastbox(3, $g.________testlib.basictypes.Integer));
            $current = 7;
            return;

          case 7:
            $current = 4;
            $continue($yield, $yieldin, $reject, $done);
            return;

          default:
            $done();
            return;
        }
      }
    };
    return $generator.new($continue, false, $g.________testlib.basictypes.Integer);
  };
  $static.SomeGenerator = function () {
    var $current = 0;
    var $continue = function ($yield, $yieldin, $reject, $done) {
      while (true) {
        switch ($current) {
          case 0:
            $yieldin($g.nested.AnotherGenerator());
            $current = 1;
            return;

          case 1:
            $yield($t.fastbox(5, $g.________testlib.basictypes.Integer));
            $current = 2;
            return;

          default:
            $done();
            return;
        }
      }
    };
    return $generator.new($continue, true, $g.________testlib.basictypes.Integer);
  };
  $static.TEST = $t.markpromising(function () {
    var $result;
    var $temp0;
    var $temp1;
    var v;
    var value;
    var $current = 0;
    var $continue = function ($resolve, $reject) {
      localasyncloop: while (true) {
        switch ($current) {
          case 0:
            v = $t.fastbox(0, $g.________testlib.basictypes.Integer);
            $current = 1;
            continue localasyncloop;

          case 1:
            $temp1 = $g.nested.SomeGenerator();
            $current = 2;
            continue localasyncloop;

          case 2:
            $promise.maybe($temp1.Next()).then(function ($result0) {
              $temp0 = $result0;
              $result = $temp0;
              $current = 3;
              $continue($resolve, $reject);
              return;
            }).catch(function (err) {
              $reject(err);
              return;
            });
            return;

          case 3:
            value = $temp0.First;
            if ($temp0.Second.$wrapped) {
              $current = 4;
              continue localasyncloop;
            } else {
              $current = 5;
              continue localasyncloop;
            }
            break;

          case 4:
            v = $t.fastbox(v.$wrapped + value.$wrapped, $g.________testlib.basictypes.Integer);
            $current = 2;
            continue localasyncloop;

          case 5:
            $resolve($t.fastbox(v.$wrapped == 12, $g.________testlib.basictypes.Boolean));
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
