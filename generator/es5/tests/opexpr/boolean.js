$module('boolean', function () {
  var $static = this;
  $static.TEST = function () {
    var first;
    var second;
    first = $t.fastbox(true, $g.________testlib.basictypes.Boolean);
    second = $t.fastbox(false, $g.________testlib.basictypes.Boolean);
    return $t.fastbox(((first.$wrapped && second.$wrapped) || first.$wrapped) || !second.$wrapped, $g.________testlib.basictypes.Boolean);
  };
});
