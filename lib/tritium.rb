
require 'nokogiri'

module Tritium
end

require_relative 'tritium/config'
require_relative 'tritium/extensions/regexp'
require_relative 'tritium/engines/legacy/legacy_base'
require_relative 'tritium/engines/legacy/debug/engine'
require_relative 'tritium/engines/legacy/reference/engine'
require_relative 'tritium/engines/legacy/tidy/engine'