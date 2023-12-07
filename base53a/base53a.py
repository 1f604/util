__all__ = ["Base53ID", "B53_generate_random_Base53ID", "b53_generate_next_Base53ID"]

######### basic checks ##########
import sys, string, random
assert sys.version_info >= (3, 11), "Use Python 3.11 or newer"

def _isprime(n:int):
	if n < 2:
		return False
	for i in range(2,n):
		if n % i == 0:
			return False
	return True

illegal_characters = set("O91IlWwmd")
illegal_pairs = set(["VV", "vv", "rn", "nn"])
legal_chars = set(string.ascii_letters + string.digits) - illegal_characters
alphabet = list(legal_chars)

p = len(alphabet)
assert _isprime(p), "size of alphabet is not prime"
assert p == 53, "size of alphabet is not 53"
assert p == len(set(alphabet)), "alphabet contains duplicates"

######### end of basic checks ############

alphabet_without_v = list(legal_chars - set(['v']))
alphabet_without_V = list(legal_chars - set(['V']))
alphabet_without_n = list(legal_chars - set(['n']))
remapping_table = {'O':'0', '9':'g'}
ordered_alphabet = sorted(alphabet)
char_to_num = {ordered_alphabet[i]:i for i in range(len(alphabet))}
num_to_char = {v:k for k,v in char_to_num.items()}


class ValidationResult(object):
	def __init__(self, success, message=''):
		assert type(success) is bool
		self.success = success
		self.message = message

	def __eq__(self, other):
		assert type(other) is ValidationResult
		return self.success == other.success and self.message == other.message

	def __bool__(self):
		return self.success

	def __repr__(self):
		return f"ValidationResult({self.success}, '{self.message}')"


class Base53ID(object):
	# construction is validation
	def __init__(self, *, string_without_checksum:str, checksum_char:str, remap:bool=False):
		# 2. check lengths
		assert len(checksum_char) == 1, ValidationResult(False, 'Checksum must be exactly one character')
		assert len(string_without_checksum) >= 1, ValidationResult(False, 'String too short')
		assert len(string_without_checksum) <= p-2, ValidationResult(False, 'String too long')
		# 2a. Remap if remapping is specified. By default we do not remap.
		if remap:
			string_without_checksum = _b53a_remap(string_without_checksum)
			checksum_char = _b53a_remap(checksum_char)
		# 3. check no illegal characters or pairs in the entire string including checksum character
		validation_result = _b53a_check_for_illegal_chars_and_pairs(string_without_checksum + checksum_char)
		assert validation_result, validation_result
		# 4. check the checksum
		# recalculate the checksum
		recalculated_checksum = _b53a_internal_get_checksum(string_without_checksum)
		assert checksum_char == recalculated_checksum, ValidationResult(False, "Checksum does not match")
		# everything checks out
		self.str_without_csum = string_without_checksum
		self.checksum = checksum_char

	def __repr__(self) -> str:
		return f"Base53ID(string_without_checksum='{self.str_without_csum}', checksum_char='{self.checksum}')"

	def __str__(self):
		return self.str_without_csum + self.checksum

	def __eq__(self, other):
		if type(other) is not type(self): # it's stupid that I have to explicitly specify this...
			return False
		return str(self) == str(other)

	def __hash__(self):
		return hash(str(self))


def B53_generate_random_Base53ID(n:int) -> Base53ID:
	for _ in range(50):
		string_without_checksum = _b53a_generate_random_unchecksummed(n)
		checksum = _b53a_internal_get_checksum(string_without_checksum)
		if all(ip not in string_without_checksum+checksum for ip in illegal_pairs):
			break
	else:
		raise Exception("This shouldn't happen!!!")
	return Base53ID(string_without_checksum=string_without_checksum, checksum_char=checksum)


def b53_generate_next_Base53ID(old_id:Base53ID) -> Base53ID:
	# length of output shall be equal to or greater than length of input
	# 00 -> 02, 002 -> 003, 0004 -> 0005
	# special case rollover: z -> 00, zz -> 000, zzz -> 0000
	assert type(old_id) is Base53ID
	# calculate the numerical equivalent of the ID
	new_s = old_id.str_without_csum
	# if all z's, then roll over to the next 0s
	if all(c == 'z' for c in old_id.str_without_csum):
		new_s = '0' * (len(old_id.str_without_csum) + 1)
		checksum = _b53a_internal_get_checksum(new_s)
		return Base53ID(string_without_checksum=new_s, checksum_char=checksum)

	for _ in range(5): # try 5 times before giving up
		# generate the next string
		new_s = _increment_base53_string(new_s)
		# check for forbidden pairs in the string_without_checksum
		try:
            checksum_char = _b53a_internal_get_checksum(string_without_checksum)
			new_id = Base53ID(string_without_checksum=new_s, checksum_char=checksum_char)
			return new_id
		except Exception as e: # the only expected exception is forbidden pair
			vr = e.args[0]
			assert type(vr) is ValidationResult, vr
			assert vr.success == False
			assert vr.message.startswith('Error: illegal pair:')
			# now we extract it
			# it should look something like this:
			# hagsf7465vv00000000000
			# the disallowed pairs are: ["VV", "vv", "rn", "nn"]
			# which does not include 'zz'
			# this is good because only 'zz' rolls over to '000'
			# so incrementing an illegal pair will still result in a pair, not a triple
			# so we can edit just the illegal pair in the string without changing the rest of the string
			new_arr = list(new_s)
			for i in range(len(new_s)-1): # find the illegal pair
				pair = new_s[i] + new_s[i+1]
				if pair in illegal_pairs:
					next_pair = _increment_base53_string(pair)
					new_arr[i] = next_pair[0]
					new_arr[i+1] = next_pair[1]
			new_s = ''.join(new_arr)

		checksum = _b53a_internal_get_checksum(new_s)
		# check with checksum
		try:
			new_id = Base53ID(string_without_checksum=new_s, checksum_char=checksum)
			return new_id
		except Exception as e:
			pass
	else:
		raise Exception("Unable to generate new ID. This should never happen.", old_id)

############################# internal implementation functions ####################################

def _increment_base53_string(str_without_csum:str) -> str:
	n = len(str_without_csum)
	s = str_without_csum
	# compute the numerical sum
	total = 0
	for i,c in enumerate(reversed(s)):
		total += char_to_num[c] * p ** i
	total += 1
	# convert the sum back into characters
	new_str = []
	power = 0
	while total:
		power += 1
		total, remainder = divmod(total, p)
		new_str.append(num_to_char[remainder])
	# prepend 0s as necessary
	diff = n - len(new_str)
	if diff > 0:
		new_str.extend(['0']*diff)
	return ''.join(reversed(new_str))


def _b53a_remap(s:str):
	ls = list(s)
	for i in range(len(ls)):
		if ls[i] in remapping_table:
			ls[i] = remapping_table[ls[i]]
	return ''.join(ls)


def _b53a_generate_random_unchecksummed(n:int) -> str:
	prev_char = ''
	result = []
	for i in range(n):
		if prev_char == 'v':
			choices = alphabet_without_v
		elif prev_char == 'V':
			choices = alphabet_without_V
		elif prev_char in ('n', 'r'):
			choices = alphabet_without_n
		else:
			choices = alphabet
		prev_char = random.choice(choices)
		result.append(prev_char)
	return ''.join(result)


def _b53a_check_for_illegal_chars_and_pairs(s:str) -> ValidationResult:
	# check for illegal characters
	for c in s:
		if c not in alphabet:
			return ValidationResult(False, f"Error: illegal character: {c}")
	# check for illegal pairs
	for ip in illegal_pairs:
		if ip in s:
			return ValidationResult(False, f"Error: illegal pair: {ip}")
	return ValidationResult(True)


def _b53a_internal_get_checksum(s:str):
	# calculate the checksum of the string
	total = 0
	for i, c in enumerate(s,1):
		multiplier = p-1-i
		num = char_to_num[c]
		total += multiplier * num
	checksum = total % p
	return num_to_char[checksum]
