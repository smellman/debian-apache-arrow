# -*- ruby -*-
#
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

require_relative "lib/arrow/version"

Gem::Specification.new do |spec|
  spec.name = "red-arrow"
  version_components = [
    Arrow::Version::MAJOR.to_s,
    Arrow::Version::MINOR.to_s,
    Arrow::Version::MICRO.to_s,
    Arrow::Version::TAG,
  ]
  spec.version = version_components.compact.join(".")
  spec.homepage = "https://arrow.apache.org/"
  spec.authors = ["Apache Arrow Developers"]
  spec.email = ["dev@arrow.apache.org"]

  spec.summary = "Red Arrow is the Ruby bindings of Apache Arrow"
  spec.description =
    "Apache Arrow is a common in-memory columnar data store. " +
    "It's useful to share and process large data."
  spec.license = "Apache-2.0"
  spec.files = ["README.md", "Rakefile", "Gemfile", "#{spec.name}.gemspec"]
  spec.files += ["LICENSE.txt", "NOTICE.txt"]
  spec.files += Dir.glob("ext/**/*.{cpp,hpp,rb}")
  spec.files += Dir.glob("lib/**/*.rb")
  spec.files += Dir.glob("image/*.*")
  spec.files += Dir.glob("doc/text/*")
  spec.test_files += Dir.glob("test/**/*")
  spec.extensions = ["ext/arrow/extconf.rb"]

  spec.add_runtime_dependency("bigdecimal", ">= 2.0.3")
  spec.add_runtime_dependency("extpp", ">= 0.0.7")
  spec.add_runtime_dependency("gio2", ">= 3.5.0")
  spec.add_runtime_dependency("native-package-installer")
  spec.add_runtime_dependency("pkg-config")

  required_msys2_package_version = version_components[0, 3].join(".")
  spec.metadata["msys2_mingw_dependencies"] =
    "arrow>=#{required_msys2_package_version}"
end
