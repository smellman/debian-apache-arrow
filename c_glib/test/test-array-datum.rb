# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.

class TestArrayDatum < Test::Unit::TestCase
  include Helper::Buildable

  def setup
    @array = build_boolean_array([true, false])
    @datum = Arrow::ArrayDatum.new(@array)
  end

  def test_array?
    assert do
      @datum.array?
    end
  end

  def test_array_like?
    assert do
      @datum.array_like?
    end
  end

  def test_scalar?
    assert do
      not @datum.scalar?
    end
  end

  def test_value?
    assert do
      @datum.value?
    end
  end

  sub_test_case("==") do
    def test_true
      assert_equal(Arrow::ArrayDatum.new(@array),
                   Arrow::ArrayDatum.new(@array))
    end

    def test_false
      table = build_table("visible" => @array)
      assert_not_equal(@datum,
                       Arrow::TableDatum.new(table))
    end
  end

  def test_to_string
    assert_equal("Array([\n" + "  true,\n" + "  false\n" + "])", @datum.to_s)
  end

  def test_value
    assert_equal(@array, @datum.value)
  end
end
